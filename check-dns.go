package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
)

func getDnsServer(plan Plan) []string {
	fmt.Printf("\n%sFetching DNS servers for %s plan...\n", colorMap["blue"], plan.String())
	// get the DNS server from shecan.ir/dns/{plan}.txt and return
	url := fmt.Sprintf("https://shecan.ir/dns/%s.txt", strings.ToLower(plan.String()))

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return []string{}
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return []string{}
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return []string{}
	}

	return unique(strings.Split(string(body), "\n"))

}

func checkDNS(plan Plan) []string {
	// get the DNS servers
	dnsServers := getDnsServer(plan)

	fmt.Printf("\n%sChecking DNS servers...\n", colorMap["blue"])
	for _, server := range dnsServers {
		if server == "" {
			continue
		}
		// ping the server for 4 times and return the average time
		// print the server and the average time
		ping := Ping(server, 4, 2)
		if report.PingReports == nil {
			report.PingReports = make(map[string]string)
		}
		report.PingReports[server] = fmt.Sprintf("%.2f ms", ping)
	}

	return dnsServers
}

func disableIPv6() {
	switch runtime.GOOS {
	case "linux":
		disableIPv6Linux()
	case "darwin":
		disableIPv6Mac()
	case "windows":
		disableIPv6Windows()
	default:
		log.Println("Unsupported OS:", runtime.GOOS)
	}
}

func disableIPv6Linux() {
	cmd := exec.Command("sysctl", "-w", "net.ipv6.conf.all.disable_ipv6=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Linux: Failed to disable IPv6:", err)
	} else {
		log.Println("Linux: IPv6 disabled:", string(out))
	}
}

func disableIPv6Mac() {
	// Replace "Wi-Fi" with your actual interface if needed
	cmd := exec.Command("networksetup", "-setv6off", "Wi-Fi")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("macOS: Failed to disable IPv6:", err)
	} else {
		log.Println("macOS: IPv6 disabled on Wi-Fi:", string(out))
	}
}

func disableIPv6Windows() {
	cmd := exec.Command("reg", "add", `HKLM\SYSTEM\CurrentControlSet\Services\Tcpip6\Parameters`, "/v", "DisabledComponents", "/t", "REG_DWORD", "/d", "0xffffffff", "/f")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Windows: Failed to disable IPv6:", err)
	} else {
		log.Println("Windows: IPv6 disabled via registry (reboot needed):", string(out))
	}
}
