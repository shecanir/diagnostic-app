package main

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

const maxHTTPConcurrency = 4

var (
	requestResultMu sync.Mutex
	checkShecanMu   sync.Mutex
)

func recordRequestResult(domain, value string) {
	requestResultMu.Lock()
	defer requestResultMu.Unlock()
	report.RequestResult[domain] = value
}

func performShecanDomainChecks(domains []string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxHTTPConcurrency)

	for _, rawDomain := range domains {
		domain := strings.TrimSpace(rawDomain)
		if domain == "" {
			continue
		}

		wg.Add(1)
		go func(d string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			response, err := HTTPRequest("https://"+d, "GET", "", "", "2")
			if err != nil {
				fmt.Println(colorMap["red"], "[Error] Can't Get", d)
				recordRequestResult(d, fmt.Sprintf("Error: %v", err))
				return
			}
			defer response.Body.Close()

			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println(colorMap["red"], "[Error] Can't Read", d)
				recordRequestResult(d, fmt.Sprintf("Error reading body: %v", err))
				return
			}

			fmt.Println(colorMap["blue"], "[INFO] Response:", string(body))
			recordRequestResult(d, string(body))
		}(domain)
	}

	wg.Wait()
}

func recordCheckShecanResult(ip string, entry CheckShecan) {
	checkShecanMu.Lock()
	defer checkShecanMu.Unlock()
	report.CheckShecanResult[ip] = entry
}

func performShecanOverIPChecks(ips []string) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, maxHTTPConcurrency)

	for _, rawIP := range ips {
		ip := strings.TrimSpace(rawIP)
		if ip == "" || isHTTPReachable(ip) {
			continue
		}

		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fmt.Println(colorMap["blue"], "[INFO] Checking Shecan Over IP:", target)
			response, err := HTTPRequest(fmt.Sprintf("https://%s", target), "GET", "", "Host: check.shecan.ir")
			if err != nil {
				fmt.Println(colorMap["red"], "[Error] Can't Get Check Shecan Result")
				fmt.Println(colorMap["red"], err)
				recordCheckShecanResult(target, CheckShecan{Error: fmt.Sprintf("Error: %v", err)})
				return
			}
			defer response.Body.Close()

			markHTTPReachable(target)
			body, err := io.ReadAll(response.Body)
			if err != nil {
				fmt.Println(colorMap["red"], "[Error] Can't Read Check Shecan Result")
				recordCheckShecanResult(target, CheckShecan{Error: fmt.Sprintf("Error reading body: %v", err), Code: response.StatusCode})
				return
			}

			fmt.Println(colorMap["blue"], "[INFO] Check Shecan Result:", string(body))
			recordCheckShecanResult(target, CheckShecan{Result: string(body), Code: response.StatusCode})
		}(ip)
	}

	wg.Wait()
}
