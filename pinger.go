package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

const maxPingConcurrency = 4

var pingMu sync.Mutex

func pingServer(server string, count int, timeout int) (float64, error) {

	var cmd *exec.Cmd

	// Choose appropriate ping command based on OS
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", strconv.Itoa(count), "-w", strconv.Itoa(timeout*1000), server)
	default: // Linux & macOS
		cmd = exec.Command("ping", "-c", strconv.Itoa(count), "-W", strconv.Itoa(timeout), server)
	}

	// Run the command and capture output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running ping command:", err)
		return -1, err
	}

	output := out.String()
	return extractAvgRTT(output)
}

// Extracts the average round-trip time (RTT) from ping output
func extractAvgRTT(output string) (float64, error) {
	// Regex patterns for different OS outputs
	patterns := []string{
		`Average = (\d+)ms`, // Windows
		`min/avg/max/stddev = [\d.]+/([\d.]+)/[\d.]+/(?:[\d.]+|nan) ms`, // Linux/macOS (handles 'nan')
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			rtt, err := strconv.ParseFloat(strings.TrimSpace(matches[1]), 64)
			if err == nil {
				return rtt, nil
			}
		}
	}

	return -1, fmt.Errorf("could not parse ping output")
}

func Ping(server string, count int, timeout int) float64 {
	fmt.Printf("%sPinging %s...\n", colorMap["green"], server)
	ping, err := pingServer(server, count, timeout)
	if err != nil {
		fmt.Println(colorMap["red"], "[Error] Error pinging server:", err)
	}
	var color string
	if ping > 600 || ping == -1 {
		color = colorMap["red"] + "❌ "
	} else {
		color = colorMap["grey"] + "✅ "
	}
	fmt.Printf("%sAvg RTT: %.2f ms\n", color, ping)
	recordPingResult(server, ping)
	return ping
}

func recordPingResult(server string, ping float64) {
	pingMu.Lock()
	defer pingMu.Unlock()
	if report.PingReports == nil {
		report.PingReports = make(map[string]string)
	}
	report.PingReports[server] = fmt.Sprintf("%.2f ms", ping)
}

func hostAlreadyPinged(server string) bool {
	pingMu.Lock()
	defer pingMu.Unlock()
	if report.PingReports == nil {
		return false
	}
	_, ok := report.PingReports[server]
	return ok
}
