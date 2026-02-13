package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

// RequestError represents custom error types for HTTP requests
type RequestError struct {
	StatusCode int
	Err        error
}

func (e *RequestError) Error() string {
	return fmt.Sprintf("request failed with status %d: %v", e.StatusCode, e.Err)
}

// RequestConfig holds the configuration for HTTP requests
type RequestConfig struct {
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

// DefaultConfig provides sensible default values
var DefaultConfig = RequestConfig{
	Timeout:    10 * time.Second,
	MaxRetries: 3,
	RetryDelay: 100 * time.Millisecond,
}

var sharedCookieJar http.CookieJar

func init() {
	if jar, err := cookiejar.New(nil); err == nil {
		sharedCookieJar = jar
	}
}

// HTTPRequest sends an HTTP request with the specified URL, optional method, body, and headers.
func HTTPRequest(url string, rest ...string) (*http.Response, error) {
	return HTTPRequestWithContext(context.Background(), url, rest...)
}

// HTTPRequestWithContext sends an HTTP request with context and retry support
func HTTPRequestWithContext(ctx context.Context, url string, rest ...string) (*http.Response, error) {
	config := DefaultConfig

	// Parse timeout if provided
	if len(rest) > 3 && rest[3] != "" {
		if timeout, err := strconv.Atoi(rest[3]); err == nil {
			config.Timeout = time.Duration(timeout) * time.Second
		}
	}

	// Extract host override for SNI and Host header
	var overrideHost string
	if len(rest) > 2 {
		headers := strings.Split(rest[2], ",")
		for _, header := range headers {
			headerParts := strings.SplitN(header, ":", 2)
			if len(headerParts) == 2 && strings.ToLower(strings.TrimSpace(headerParts[0])) == "host" {
				overrideHost = strings.TrimSpace(headerParts[1])
				break
			}
		}
	}

	client := newHTTPClient(config.Timeout, overrideHost)

	var resp *http.Response
	var lastErr error
	challengeRetried := false

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * config.RetryDelay)
			}

			req, err := prepareRequest(ctx, url, rest...)
			if err != nil {
				return nil, fmt.Errorf("failed to prepare request: %w", err)
			}

			if overrideHost != "" {
				req.Host = overrideHost
			}

			resp, lastErr = client.Do(req)
			if lastErr == nil {
				if shouldRetryForChallenge(resp, req.URL.Hostname(), challengeRetried) {
					challengeRetried = true
					discardResponse(resp)
					attempt--
					continue
				}

				if resp.StatusCode < 500 {
					return resp, nil
				}
			}

			if resp != nil {
				resp.Body.Close()
			}
		}
	}

	return nil, &RequestError{
		StatusCode: getStatusCode(resp),
		Err:        lastErr,
	}
}

func prepareRequest(ctx context.Context, url string, rest ...string) (*http.Request, error) {
	method := "GET"
	if len(rest) > 0 && rest[0] != "" {
		method = rest[0]
	}

	var body io.Reader
	if len(rest) > 1 && rest[1] != "" {
		body = strings.NewReader(rest[1])
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if len(rest) > 2 {
		headers := strings.Split(rest[2], ",")
		for _, header := range headers {
			headerParts := strings.SplitN(header, ":", 2)
			if len(headerParts) == 2 {
				req.Header.Set(strings.TrimSpace(headerParts[0]), strings.TrimSpace(headerParts[1]))
			}
		}
	} else {
		req.Header.Set("User-Agent", "Go-HTTP-Client")
		req.Header.Set("Accept", "application/json")
	}

	return req, nil
}

func getStatusCode(resp *http.Response) int {
	if resp == nil {
		return http.StatusInternalServerError
	}
	return resp.StatusCode
}

func newHTTPClient(timeout time.Duration, overrideHost string) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			ServerName: overrideHost,
			// InsecureSkipVerify: true, // Uncomment if using self-signed certs
		},
		DialContext: (&net.Dialer{
			Timeout:   timeout,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	if sharedCookieJar != nil {
		client.Jar = sharedCookieJar
	}

	return client
}

func shouldRetryForChallenge(resp *http.Response, host string, alreadyRetried bool) bool {
	if resp == nil || alreadyRetried {
		return false
	}
	host = strings.ToLower(strings.TrimSpace(strings.Split(host, ":")[0]))
	if !isShecanHost(host) {
		return false
	}

	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusServiceUnavailable, http.StatusTooManyRequests:
	default:
		return false
	}

	if len(resp.Cookies()) == 0 && len(resp.Header["Set-Cookie"]) == 0 {
		return false
	}

	return true
}

func isShecanHost(host string) bool {
	if host == "" {
		return false
	}
	return host == "shecan.ir" || strings.HasSuffix(host, ".shecan.ir")
}

func discardResponse(resp *http.Response) {
	if resp == nil || resp.Body == nil {
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}
