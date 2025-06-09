package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// DNSRecord represents a single DNS query result, including resolver and address details
type DNSRecord struct {
	Domain   string `json:"domain"`
	Resolver string `json:"resolver"`
	Address  string `json:"address"`
	Value    string `json:"value"` // Updated field for resolved IP
	Resolved string `json:"resolved_at"`
	Error    string `json:"error,omitempty"` // Stores errors if the lookup fails
}

// RunCommand executes a shell command with a timeout
func RunCommand(timeout time.Duration, command string, args ...string) (string, error) {
	fmt.Println(colorMap["blue"], "[INFO] Running command:", command, args, colorMap["reset"])

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, command, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println(colorMap["red"], "[ERROR] Command timed out", colorMap["reset"])
		return "", fmt.Errorf("command timed out")
	}

	if err != nil {
		fmt.Println(colorMap["red"], "[ERROR] Command execution failed:", err, colorMap["reset"])
	}
	return out.String(), err
}

// ParseNslookupOutput parses `nslookup` output and extracts resolver, address, and errors
func ParseNslookupOutput(output, domain string) ([]DNSRecord, error) {
	fmt.Println(colorMap["blue"], "[INFO] Parsing nslookup output for domain:", domain, colorMap["reset"])

	var records []DNSRecord
	lines := strings.Split(output, "\n")

	var foundResolver string
	var resolvedValues []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Server:") {
			foundResolver = strings.Fields(line)[1] // Extract resolver name or IP
			fmt.Println(colorMap["green"], "[INFO] Detected resolver:", foundResolver, colorMap["reset"])
		}
		if strings.Contains(line, "Address:") && !strings.HasPrefix(line, "Server:") {
			fields := strings.Fields(line)
			if len(fields) > 1 {
				fmt.Println(colorMap["green"], "[INFO] Found resolved address:", fields[1], colorMap["reset"])
				resolvedValues = append(resolvedValues, fields[1])
			}
		}
	}

	if len(resolvedValues) == 0 {
		fmt.Println(colorMap["red"], "[ERROR] Failed to parse nslookup output", colorMap["reset"])
		return nil, fmt.Errorf("failed to parse nslookup output: %s", output)
	}

	// The first resolved value is assumed to be the resolver address
	if foundResolver == "" && len(resolvedValues) > 0 {
		foundResolver = resolvedValues[0]
		resolvedValues = resolvedValues[1:] // Remove resolver from resolved values
	}

	// Store only the first resolved value
	if len(resolvedValues) > 0 {
		records = append(records, DNSRecord{
			Domain:   domain,
			Resolver: foundResolver,
			Address:  fmt.Sprintf("%s#53", foundResolver),
			Value:    resolvedValues[0], // Assign resolved value correctly
			Resolved: time.Now().Format(time.RFC3339),
		})
	}

	fmt.Println(colorMap["blue"], "[INFO] Successfully parsed", len(records), "records for", domain, colorMap["reset"])
	return records, nil
}

// NsLookup performs a DNS lookup using `nslookup`
func NsLookup(domain string) []DNSRecord {
	const timeout = 5 * time.Second
	fmt.Println(colorMap["blue"], "[INFO] Querying DNS for domain:", domain, colorMap["reset"])

	cmdOutput, err := RunCommand(timeout, "nslookup", domain)
	if err != nil {
		fmt.Println(colorMap["red"], "[ERROR] nslookup command failed for domain:", domain, colorMap["reset"])
		return []DNSRecord{{
			Domain:   domain,
			Resolver: "",
			Address:  "",
			Value:    "",
			Resolved: time.Now().Format(time.RFC3339),
			Error:    err.Error(),
		}}
	}

	parsedRecords, parseErr := ParseNslookupOutput(cmdOutput, domain)
	if parseErr != nil {
		fmt.Println(colorMap["red"], "[ERROR] Parsing failed for domain:", domain, colorMap["reset"])
		return []DNSRecord{{
			Domain:   domain,
			Resolver: "",
			Address:  "",
			Value:    "",
			Resolved: time.Now().Format(time.RFC3339),
			Error:    parseErr.Error(),
		}}
	}

	fmt.Println(colorMap["blue"], "[INFO] Successfully queried DNS for", domain, colorMap["reset"])
	return parsedRecords
}
