package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Result represents the scan result for a single domain
type Result struct {
	Domain     string `json:"domain"`
	Alive      bool   `json:"alive"`
	StatusCode int    `json:"status_code,omitempty"`
	Scheme     string `json:"scheme,omitempty"`
	Server     string `json:"server,omitempty"`
	Error      string `json:"error,omitempty"`
}

func checkDomain(domain string, timeout time.Duration) Result {
	client := http.Client{
		Timeout: timeout,
		// Skip certificate verification since the goal is fast discovery
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		// Don't follow more than 5 redirects
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// Try https first, then fall back to http
	schemes := []string{"https", "http"}
	var lastErr error

	for _, scheme := range schemes {
		url := fmt.Sprintf("%s://%s", scheme, domain)
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		server := resp.Header.Get("Server")
		return Result{
			Domain:     domain,
			Alive:      true,
			StatusCode: resp.StatusCode,
			Scheme:     scheme,
			Server:     server,
		}
	}

	errMsg := "connection failed"
	if lastErr != nil {
		errMsg = lastErr.Error()
	}
	return Result{
		Domain: domain,
		Alive:  false,
		Error:  errMsg,
	}
}

func readDomains(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var domains []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		domains = append(domains, line)
	}
	return domains, scanner.Err()
}

func worker(id int, jobs <-chan string, results chan<- Result, timeout time.Duration, wg *sync.WaitGroup) {
	defer wg.Done()
	for domain := range jobs {
		results <- checkDomain(domain, timeout)
	}
}

func exportJSON(results []Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}

func exportCSV(results []Result, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	defer w.Flush()

	w.Write([]string{"domain", "alive", "status_code", "scheme", "server", "error"})
	for _, r := range results {
		status := ""
		if r.StatusCode != 0 {
			status = fmt.Sprintf("%d", r.StatusCode)
		}
		alive := "false"
		if r.Alive {
			alive = "true"
		}
		w.Write([]string{r.Domain, alive, status, r.Scheme, r.Server, r.Error})
	}
	return nil
}

func main() {
	inputFile := flag.String("l", "", "File containing the list of domains (one per line) - required")
	concurrency := flag.Int("c", 50, "Number of concurrent scans (Concurrency)")
	timeoutSec := flag.Int("t", 5, "Timeout in seconds for each request")
	outJSON := flag.String("json", "", "Path to export results as JSON (optional)")
	outCSV := flag.String("csv", "", "Path to export results as CSV (optional)")
	showOnlyAlive := flag.Bool("alive-only", false, "Show only alive domains in the console")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("Error: you must specify a domain file using -l")
		fmt.Println("Example: fastsub-scanner -l domains.txt -c 100 -json results.json")
		os.Exit(1)
	}

	domains, err := readDomains(*inputFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	if len(domains) == 0 {
		fmt.Println("The file is empty or contains no valid domains.")
		os.Exit(1)
	}

	fmt.Printf("[*] Loaded %d domain(s) to scan\n", len(domains))
	fmt.Printf("[*] Concurrency: %d | Timeout: %ds\n\n", *concurrency, *timeoutSec)

	start := time.Now()

	jobs := make(chan string, len(domains))
	resultsCh := make(chan Result, len(domains))
	var wg sync.WaitGroup

	// Launch the worker pool
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(i, jobs, resultsCh, time.Duration(*timeoutSec)*time.Second, &wg)
	}

	// Feed the channel with domains
	for _, d := range domains {
		jobs <- d
	}
	close(jobs)

	// Close the results channel once all workers are done
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var results []Result
	aliveCount := 0

	for r := range resultsCh {
		results = append(results, r)
		if r.Alive {
			aliveCount++
			fmt.Printf("[Alive] %-35s | %3d | %-6s | Server: %s\n", r.Domain, r.StatusCode, r.Scheme, orDash(r.Server))
		} else if !*showOnlyAlive {
			fmt.Printf("[Down]  %-35s | %s\n", r.Domain, r.Error)
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("\n[*] Scan completed in %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("[*] Result: %d alive / %d total\n", aliveCount, len(domains))

	if *outJSON != "" {
		if err := exportJSON(results, *outJSON); err != nil {
			fmt.Printf("Error exporting JSON: %v\n", err)
		} else {
			fmt.Printf("[+] Results saved to: %s\n", *outJSON)
		}
	}

	if *outCSV != "" {
		if err := exportCSV(results, *outCSV); err != nil {
			fmt.Printf("Error exporting CSV: %v\n", err)
		} else {
			fmt.Printf("[+] Results saved to: %s\n", *outCSV)
		}
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
