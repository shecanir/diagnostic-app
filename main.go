package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Plan type representing different diagnostic plans
type Plan int

const (
	Free Plan = iota + 1 // 1
	Pro                  // 2
)

// String method to convert enum values to strings
func (p Plan) String() string {
	switch p {
	case Free:
		return "Free"
	case Pro:
		return "Pro"
	default:
		return "Unknown"
	}
}

func parsePlan(input string) Plan {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "free":
		return Free
	case "pro":
		return Pro
	default:
		return Pro // default to Pro
	}
}

var (
	report Report = initReport()
)

// MarshalJSON converts the Plan enum to a JSON string
func (p Plan) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON converts a JSON string to a Plan enum
func (p *Plan) UnmarshalJSON(data []byte) error {
	var planStr string
	if err := json.Unmarshal(data, &planStr); err != nil {
		return err
	}

	switch planStr {
	case "Free":
		*p = Free
	case "Pro":
		*p = Pro
	default:
		*p = Free // Defaulting to Free if an unknown value is encountered
	}
	return nil
}

func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Print(colorMap["yellow"], "[Warning] .env file not found, using default values", colorMap["reset"])
	}

	if err := Execute(); err != nil {
		fmt.Println(colorMap["red"], "[Error]", err)
		os.Exit(1)
	}
}

func runDiagnostic() {
	// print the logo
	printLogo()

	var selectedPlan Plan
	reader := bufio.NewReader(os.Stdin)
	if PlanFlag != "" {
		selectedPlan = parsePlan(PlanFlag)
	} else {
		fmt.Println("Select the plan to diagnose:")
		fmt.Printf("%d. %s\n", Free, Free)
		fmt.Printf("%d. %s\n", Pro, Pro)

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter your choice (1 or 2, default is Pro): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		selectedPlan = Pro // Default plan is Pro
		switch input {
		case "2", "":
			selectedPlan = Pro
		case "1":
			selectedPlan = Free
		}
	}
	report.Plan = selectedPlan

	// get shecan DNS servers based on the selected plan
	shecanDNS := checkDNS(selectedPlan)

	// check os DNS servers if shecan not set return error check with report.DNSServers
	if len(shecanDNS) == 0 {
		fmt.Println(colorMap["red"], "[Error] Can't Get Shecan DNS")
		return
	} else {

		// os DNS comparison temporarily disabled
	}

	// if pro plan selected ask to get updaterLink and store it in report.UpdaterLink
	if selectedPlan == Pro {
		fmt.Print("Enter the updater link: (default is empty): ")
		updaterLink, _ := reader.ReadString('\n')
		updaterLink = strings.TrimSpace(updaterLink)
		// check updater link has pattern of ^https:\/\/ddns\.shecan\.ir\/update\?password\=[0-f]{16}$
		regexp := regexp.MustCompile(`^https:\/\/ddns\.shecan\.ir\/update\?password\=[0-f]{16}$`)
		if updaterLink != "" && !regexp.MatchString(updaterLink) {
			fmt.Println(colorMap["red"], "[Error] Invalid Updater Link, updater link should be like https://ddns.shecan.ir/update?password=[16]")
			return
		}
		report.UpdaterLink = updaterLink
	}

	// if updaterLink is not empty get the response of updaterLink and store it in report.UpdaterResponse
	if report.UpdaterLink != "" {
		response, err := HTTPRequest(report.UpdaterLink, "GET", "", "", "2")
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Get Updater Response")
			return
		}

		defer response.Body.Close()

		// convert response to string
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Read Updater Response")
			return
		}
		responseBodyStr := string(responseBody)

		if responseBodyStr == "nohost" {
			fmt.Println(colorMap["red"], "[Error] Your Order not applied yet or your password is wrong")
			return
		} else if responseBodyStr == "out of the range" {
			fmt.Println(colorMap["red"], "[Error] You Order registered as Static IP and your current IP is out of the range")
			return
		} else if responseBodyStr == "invalid" {
			fmt.Println(colorMap["red"], "[Error] Your updater link is not valid")
			return
		} else {
			// if check.shecan.ir get 403 wait for 1 minute and check again
			response, err := HTTPRequest("https://check.shecan.ir", "GET", "", "", "2")
			if err != nil {
				fmt.Println(colorMap["red"], "[Error] Can't Get Check Shecan Response")
				return
			}
			defer response.Body.Close()
			if response.StatusCode == 403 {
				// delay for 1 minute
				fmt.Println(colorMap["yellow"], "[Warning] Waiting for bit...")
				time.Sleep(1 * time.Minute)
			}
		}
	}

	// get nslookup results of [shecan.ir, check.shecan.ir, fail.shecan.ir]
	// expect fail.shecan.ir to fail
	if report.NsLookup == nil {
		report.NsLookup = make(map[string][]DNSRecord)
	}
	nslookupDomains := []string{"shecan.ir", "check.shecan.ir", "fail.shecan.ir"}

	for _, domain := range nslookupDomains {
		report.NsLookup[domain] = NsLookup(domain)
	}

	// get request to check.shecan.ir and fail.shecan.ir and store the result in report.RequestResult
	if report.RequestResult == nil {
		report.RequestResult = make(map[string]string)
	}

	fmt.Println(colorMap["blue"], "[INFO] Checking shecan domains...")
	performShecanDomainChecks(nslookupDomains[1:])

	fmt.Println(colorMap["blue"], "[INFO] Checking shecan IPs...")

	// if check.shecan.ir is not reachable or error return error
	if report.RequestResult["check.shecan.ir"] == "" || strings.Contains(report.RequestResult["check.shecan.ir"], "Error") {
		fmt.Println(colorMap["red"], "[Error] Can't Reach check.shecan.ir")
		return
	}

	// if fail.shecan.ir is reachable return error and said you are used other DNS servers, VPN, forced DNS, ...

	if report.RequestResult["fail.shecan.ir"] != "" && !strings.Contains(report.RequestResult["fail.shecan.ir"], "Error") {
		fmt.Println(colorMap["red"], "[Error] You are used other DNS servers, VPN, forced DNS, ...")
		return
	}

	// get the ips of shecan from https://check.shecan.ir/ip-list.php if not error return error and exit
	response, err := HTTPRequest("https://check.shecan.ir/ip-list.php")
	if err != nil {
		fmt.Println(colorMap["red"], "[Error] Can't Get Shecan IPs")
		return
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println(colorMap["red"], "[Error] Can't Read Shecan IPs")
		return
	}
	IPs := strings.Split(string(body), "\n")

	// remove empty strings from IPs
	var newIPs []string
	for _, ip := range IPs {
		if ip != "" {
			newIPs = append(newIPs, ip)
		}
	}
	IPs = newIPs

	if report.CheckShecanResult == nil {
		report.CheckShecanResult = make(map[string]CheckShecan)
	}

	fmt.Println(colorMap["blue"], "[INFO] Checking shecan Over IPS...")
	performShecanOverIPChecks(IPs)
	runConcurrentPings(IPs, 2, 2)
	fmt.Println(colorMap["green"], "[Success] Report Generated Successfully")
	_err := sendReport(report)
	if _err != nil {
		fmt.Println(colorMap["red"], "[Error] Can't Send Report")
		return
	}
}
