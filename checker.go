package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	green  = "\033[32m"
	red    = "\033[31m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	reset  = "\033[0m"
	border = fmt.Sprintf("%s══════════════════════════════════════════════%s", cyan, reset)
)

func main() {
	var proxiesFile, outputFile, protocol string
	startTime := time.Now()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", cyan, reset)
	fmt.Printf("%s║          Proxy Checker by m85.68           ║%s\n", yellow, reset)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", cyan, reset)
	fmt.Printf("%sEnter path to proxy file (default: proxies.txt): %s", yellow, reset)
	fmt.Scanln(&proxiesFile)
	if proxiesFile == "" {
		proxiesFile = "proxies.txt"
	}
	fmt.Printf("%sEnter path to save working proxies (default: working_proxies.txt): %s", yellow, reset)
	fmt.Scanln(&outputFile)
	if outputFile == "" {
		outputFile = "working_proxies.txt"
	}
	fmt.Printf("%sEnter proxy protocol (http, socks4, socks5, default: http): %s", yellow, reset)
	fmt.Scanln(&protocol)
	if protocol == "" {
		protocol = "http"
	}
	if protocol != "http" && protocol != "socks4" && protocol != "socks5" {
		log.Fatalf("Invalid protocol: %s. Must be http, socks4, or socks5", protocol)
	}
	file, err := os.Open(proxiesFile)
	if err != nil {
		log.Fatalf("Failed to open proxies file: %v", err)
	}
	defer file.Close()
	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()
	scanner := bufio.NewScanner(file)
	const maxBufferSize = 1024 * 1024 
	scanner.Buffer(make([]byte, maxBufferSize), maxBufferSize)
	proxyChan := make(chan string, 1000)
	var wg sync.WaitGroup
	workers := 300000 
	wg.Add(workers)
	var mu sync.Mutex
	var aliveCount, deadCount int
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 5 * time.Second,
	}
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for proxy := range proxyChan {
				status, err := checkProxy(proxy, "http://httpbin.org/ip", client, protocol)
				if err != nil || status != "ALIVE" {
					fmt.Printf("%s[DEAD]%s %s\n", red, reset, proxy)
					mu.Lock()
					deadCount++
					mu.Unlock()
				} else {
					fmt.Printf("%s[ALIVE]%s %s\n", green, reset, proxy)
					mu.Lock()
					fmt.Fprintln(outFile, proxy)
					aliveCount++
					mu.Unlock()
				}
			}
		}()
	}
	go func() {
		for scanner.Scan() {
			proxy := strings.TrimSpace(scanner.Text())
			if proxy != "" {
				proxyChan <- proxy
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading proxies: %v", err)
		}
		close(proxyChan)
	}()
	wg.Wait()
	duration := time.Since(startTime)
	total := aliveCount + deadCount
	fmt.Println()
	fmt.Printf("%s╔════════════════════════════════════════════╗%s\n", cyan, reset)
	fmt.Printf("%s║       Proxy Checking Completed!            ║%s\n", yellow, reset)
	fmt.Printf("%s║       Made by m85.68                      ║%s\n", yellow, reset)
	fmt.Printf("%s╠════════════════════════════════════════════╣%s\n", cyan, reset)
	fmt.Printf("%s║ Total Proxies : %d                        ║%s\n", green, total, reset)
	fmt.Printf("%s║ Alive         : %d                        ║%s\n", green, aliveCount, reset)
	fmt.Printf("%s║ Dead          : %d                        ║%s\n", red, deadCount, reset)
	fmt.Printf("%s║ Time Taken    : %v                        ║%s\n", green, duration.Round(time.Millisecond), reset)
	fmt.Printf("%s║ Saved to      : %s                        ║%s\n", green, outputFile, reset)
	fmt.Printf("%s╚════════════════════════════════════════════╝%s\n", cyan, reset)
}

func checkProxy(proxyStr, testURL string, client *http.Client, protocol string) (status string, err error) {
	if !strings.HasPrefix(proxyStr, "http://") && !strings.HasPrefix(proxyStr, "https://") && !strings.HasPrefix(proxyStr, "socks4://") && !strings.HasPrefix(proxyStr, "socks5://") {
		proxyStr = protocol + "://" + proxyStr
	}
	proxyURL, err := url.Parse(proxyStr)
	if err != nil {
		return "DEAD", err
	}
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}
	client.Transport = transport
	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return "DEAD", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "DEAD", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "DEAD", fmt.Errorf("status code %d", resp.StatusCode)
	}
	return "ALIVE", nil
}
