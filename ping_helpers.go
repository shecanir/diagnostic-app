package main

import (
	"strings"
	"sync"
)

var (
	httpReachableMu    sync.Mutex
	httpReachableHosts = map[string]struct{}{}
)

func markHTTPReachable(host string) {
	host = strings.TrimSpace(host)
	if host == "" {
		return
	}
	httpReachableMu.Lock()
	httpReachableHosts[host] = struct{}{}
	httpReachableMu.Unlock()
}

func isHTTPReachable(host string) bool {
	host = strings.TrimSpace(host)
	if host == "" {
		return false
	}
	httpReachableMu.Lock()
	_, ok := httpReachableHosts[host]
	httpReachableMu.Unlock()
	return ok
}

func runConcurrentPings(targets []string, count, timeout int) {
	sem := make(chan struct{}, maxPingConcurrency)
	var wg sync.WaitGroup

	for _, raw := range targets {
		host := strings.TrimSpace(raw)
		if host == "" || hostAlreadyPinged(host) || isHTTPReachable(host) {
			continue
		}

		wg.Add(1)
		go func(h string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			Ping(h, count, timeout)
		}(host)
	}

	wg.Wait()
}
