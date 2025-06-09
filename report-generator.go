package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type CheckShecan struct {
	Code   int    `json:"code"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// Report struct to hold the system information
type Report struct {
	Hostname          string                 `json:"hostname"`
	OS                string                 `json:"os"`
	IPs               []string               `json:"local_ips"`
	PublicIP          string                 `json:"public_ip"`
	Plan              Plan                   `json:"plan"`
	PingReports       map[string]string      `json:"ping_reports"`
	LocalTime         string                 `json:"local_time"`
	RealTime          string                 `json:"real_time"`
	CPUInfo           string                 `json:"cpu"`
	MemoryInfo        string                 `json:"memory"`
	DiskInfo          string                 `json:"disk"`
	DNSServers        []string               `json:"dns_servers"`
	RequestResult     map[string]string      `json:"request_result"`
	NsLookup          map[string][]DNSRecord `json:"ns_lookup"`
	CheckShecanResult map[string]CheckShecan `json:"check_shecan_result"`
	UpdaterLink       string                 `json:"updater_link"`
}

// getLocalIPs retrieves all local IPs
func getLocalIPs() ([]string, error) {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if !v.IP.IsLoopback() {
					ips = append(ips, v.IP.String())
				}
			}
		}
	}
	return ips, nil
}

// getPublicIP retrieves the external IP from shecan.ir
func getPublicIP() (string, error) {
	resp, err := http.Get("https://shecan.ir/ip/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := io.ReadAll(resp.Body) // Replaces ioutil.ReadAll
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(ip)), nil
}

// getSystemInfo gathers CPU, Memory, and Disk usage
func getSystemInfo() (string, string, string) {
	var cpuInfo, memoryInfo, diskInfo string

	switch runtime.GOOS {
	case "linux", "darwin":
		cpuInfoBytes, _ := exec.Command("sh", "-c", "sysctl -n machdep.cpu.brand_string || cat /proc/cpuinfo | grep 'model name' | uniq").Output()
		cpuInfo = strings.TrimSpace(string(cpuInfoBytes))

		memoryBytes, _ := exec.Command("sh", "-c", "vm_stat | grep 'Pages free' || free -h | grep Mem").Output()
		memoryInfo = strings.TrimSpace(string(memoryBytes))

		diskBytes, _ := exec.Command("sh", "-c", "df -h | grep '/$'").Output()
		diskInfo = strings.TrimSpace(string(diskBytes))
	case "windows":
		cpuInfo = "Windows CPU Info"
		memoryInfo = "Windows Memory Info"
		diskInfo = "Windows Disk Info"
	default:
		cpuInfo, memoryInfo, diskInfo = "Unknown", "Unknown", "Unknown"
	}

	return cpuInfo, memoryInfo, diskInfo
}

// getDNSServers retrieves the DNS servers configured in the OS
func getDNSServers() ([]string, error) {
	var dnsServers []string
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("sh", "-c", "cat /etc/resolv.conf | grep nameserver | awk '{print $2}'")
	case "darwin":
		cmd = exec.Command("sh", "-c", "scutil --dns | grep 'nameserver\\[[0-9]\\]' | awk '{print $3}'")
	case "windows":
		cmd = exec.Command("powershell", "-Command", "Get-DnsClientServerAddress | Select-Object -ExpandProperty ServerAddresses | Where-Object { $_ -match '^\\d{1,3}(\\.\\d{1,3}){3}$' }")
	default:
		return nil, fmt.Errorf("unsupported OS")
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	dnsServers = unique(lines)
	return dnsServers, nil
}

func unique(elements []string) []string {
	encountered := map[string]bool{}
	result := []string{}

	for _, v := range elements {
		if v == "" {
			continue
		}
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return result

}

// convert report to json
func (r Report) String() (string, error) {
	jsonData, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// sendReport sends the generated JSON report to a server
func sendReport(report Report) error {
	// read server URL from environment variable
	serverURL := os.Getenv("REPORT_SERVER_URL")
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	responseBody, _ := io.ReadAll(resp.Body) // Replaces ioutil.ReadAll
	fmt.Println(colorMap["reset"])
	fmt.Print(report.String())
	fmt.Println("Report Saved:", string(responseBody))
	return nil
}

func initReport() Report {
	hostname, _ := os.Hostname()
	localIPs, _ := getLocalIPs()
	publicIP, _ := getPublicIP()
	cpu, memory, disk := getSystemInfo()
	dnsServers, _ := getDNSServers()

	return Report{
		Hostname:   hostname,
		OS:         runtime.GOOS,
		IPs:        localIPs,
		PublicIP:   publicIP,
		CPUInfo:    cpu,
		DiskInfo:   disk,
		MemoryInfo: memory,
		DNSServers: dnsServers,
	}
}
