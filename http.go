package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os/exec"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/proxy"
)

var execCommand = exec.Command

var (
	clineProxyMu  sync.RWMutex
	clineProxyURL string
)

func getClineProxy() string {
	clineProxyMu.RLock()
	defer clineProxyMu.RUnlock()
	return clineProxyURL
}

func buildTransportForProxy(proxyStr string) (*http.Transport, error) {
	t := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
	}
	proxyStr = strings.TrimSpace(proxyStr)
	if proxyStr == "" {
		t.Proxy = http.ProxyFromEnvironment
		return t, nil
	}
	u, err := url.Parse(proxyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid proxy URL: %w", err)
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		t.Proxy = http.ProxyURL(u)
	case "socks5", "socks5h":
		var auth *proxy.Auth
		if u.User != nil {
			auth = &proxy.Auth{
				User: u.User.Username(),
			}
			auth.Password, _ = u.User.Password()
		}
		dialer, err := proxy.SOCKS5("tcp", u.Host, auth, proxy.Direct)
		if err != nil {
			return nil, fmt.Errorf("create socks5 dialer failed: %w", err)
		}
		t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			type dialRes struct {
				conn net.Conn
				err  error
			}
			resCh := make(chan dialRes, 1)
			go func() {
				conn, err := dialer.Dial(network, addr)
				resCh <- dialRes{conn: conn, err: err}
			}()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case r := <-resCh:
				return r.conn, r.err
			}
		}
	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", u.Scheme)
	}
	return t, nil
}

func setClineProxy(rawURL string) error {
	t, err := buildTransportForProxy(rawURL)
	if err != nil {
		return err
	}
	clineProxyMu.Lock()
	clineProxyURL = strings.TrimSpace(rawURL)
	httpClient.Transport = t
	clineProxyMu.Unlock()
	return nil
}

// testProxyConnectivity 测试代理连通性并返回 (延迟ms, http状态码, 错误)
func testProxyConnectivity(proxyStr string, targetURL string) (int64, int, error) {
	if targetURL == "" {
		targetURL = "https://api.cline.bot/api/v1/health"
	}
	t, err := buildTransportForProxy(proxyStr)
	if err != nil {
		return 0, 0, err
	}
	client := &http.Client{
		Transport: t,
		Timeout:   10 * time.Second,
	}
	start := time.Now()
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return 0, 0, err
	}
	req.Header.Set("User-Agent", "cline-proxy/1.0")
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return latency, 0, err
	}
	defer resp.Body.Close()
	return latency, resp.StatusCode, nil
}

var httpTransport = &http.Transport{
	Proxy:               http.ProxyFromEnvironment,
	MaxIdleConns:        100,
	MaxIdleConnsPerHost: 10,
	IdleConnTimeout:     90 * time.Second,
	DisableCompression:  false,
}

var httpClient = &http.Client{
	Transport: httpTransport,
}

func httpPostForm(rawURL string, form url.Values) (*http.Response, error) {
	req, err := http.NewRequest("POST", rawURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return httpClient.Do(req)
}

func httpPostJSON(rawURL string, body any) (*http.Response, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", rawURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return httpClient.Do(req)
}

func readBody(resp *http.Response) string {
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Sprintf("<read error: %v>", err)
	}
	return string(data)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}
