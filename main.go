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

		// ask to check with os DNS servers
		fmt.Print("Do you want to check with OS DNS servers? (y/n, default is n): ")
		input, _ := reader.ReadString('\n')
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "y" {

			shecanHasIPv6 := false
			// check shecanDNS has IPv6
			for _, dns := range shecanDNS {
				if strings.Contains(dns, ":") {
					shecanHasIPv6 = true
					break
				}
			}

			if report.DNSServers == nil {
				fmt.Println(colorMap["red"], "[Error] Can't Get OS DNS")
				return
			}

			// if shecan DNS servers are not equal to OS DNS servers return error
			if len(shecanDNS) != len(report.DNSServers) {
				fmt.Println(colorMap["red"], "[Error] DNS Servers are not equal", colorMap["reset"])
				return
			}
			for _, dns := range shecanDNS {
				// check dns servers are same (maybe order is different)
				if contains(report.DNSServers, dns) {
					if !shecanHasIPv6 && strings.Contains(dns, ":") {
						disableIPv6()
					}
					continue
				} else {
					fmt.Println(colorMap["red"], "[Error] DNS Servers are not equal", colorMap["reset"])
					return
				}
			}
		}
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

	for _, domain := range nslookupDomains[1:] {
		response, err := HTTPRequest("https://"+domain, "GET", "", "", "2")
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Get", domain)
			report.RequestResult[domain] = fmt.Sprintf("Error: %v", err)
			continue
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Read", domain)
			report.RequestResult[domain] = fmt.Sprintf("Error reading body: %v", err)
		} else {
			fmt.Println(colorMap["blue"], "[INFO] Response:", string(body))
			report.RequestResult[domain] = string(body)
		}
	}

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

	// ping to shecan IPs and store the result in report.PingReports
	if report.PingReports == nil {
		report.PingReports = make(map[string]string)
	}

	for _, ip := range IPs {
		ping := Ping(ip, 2, 2)
		report.PingReports[ip] = fmt.Sprintf("%.2f ms", ping)
	}

	if report.CheckShecanResult == nil {
		report.CheckShecanResult = make(map[string]CheckShecan)
	}

	fmt.Println(colorMap["blue"], "[INFO] Checking shecan Over IPS...")
	// get the result of check.shecan.ir and store it in report.CheckShecanResult
	for _, ip := range IPs {
		if ip == "" {
			continue
		}
		fmt.Println(colorMap["blue"], "[INFO] Checking Shecan Over IP:", ip)
		response, err := HTTPRequest(fmt.Sprintf("https://%s", ip), "GET", "", "Host: check.shecan.ir")
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Get Check Shecan Result")
			fmt.Println(colorMap["red"], err)
			report.CheckShecanResult[ip] = CheckShecan{Error: fmt.Sprintf("Error: %v", err)}
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			fmt.Println(colorMap["red"], "[Error] Can't Read Check Shecan Result")
			report.CheckShecanResult[ip] = CheckShecan{Error: fmt.Sprintf("Error reading body: %v", err), Code: response.StatusCode}
		} else {
			fmt.Println(colorMap["blue"], "[INFO] Check Shecan Result:", string(body))
			report.CheckShecanResult[ip] = CheckShecan{Result: string(body), Code: response.StatusCode}
		}
	}
	fmt.Println(colorMap["green"], "[Success] Report Generated Successfully")
	_err := sendReport(report)
	if _err != nil {
		fmt.Println(colorMap["red"], "[Error] Can't Send Report")
		return
	}
}
