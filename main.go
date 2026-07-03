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

// Result يمثل نتيجة فحص نطاق واحد
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
		// لا نتحقق من صحة الشهادة لأن الهدف هو سرعة الاستكشاف فقط
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		// لا تتبع أكثر من 5 تحويلات
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	// نجرب https أولاً ثم http كخطة بديلة
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
	inputFile := flag.String("l", "", "ملف يحتوي على قائمة النطاقات (سطر لكل نطاق) - إلزامي")
	concurrency := flag.Int("c", 50, "عدد الفحوصات المتزامنة (Concurrency)")
	timeoutSec := flag.Int("t", 5, "المهلة الزمنية بالثواني لكل طلب")
	outJSON := flag.String("json", "", "مسار ملف JSON لتصدير النتائج (اختياري)")
	outCSV := flag.String("csv", "", "مسار ملف CSV لتصدير النتائج (اختياري)")
	showOnlyAlive := flag.Bool("alive-only", false, "عرض النطاقات الفعّالة فقط في الطرفية")
	flag.Parse()

	if *inputFile == "" {
		fmt.Println("خطأ: يجب تحديد ملف النطاقات باستخدام -l")
		fmt.Println("مثال: fastsub-scanner -l domains.txt -c 100 -json results.json")
		os.Exit(1)
	}

	domains, err := readDomains(*inputFile)
	if err != nil {
		fmt.Printf("خطأ في قراءة الملف: %v\n", err)
		os.Exit(1)
	}

	if len(domains) == 0 {
		fmt.Println("الملف فارغ أو لا يحتوي على نطاقات صالحة.")
		os.Exit(1)
	}

	fmt.Printf("[*] تم تحميل %d نطاق للفحص\n", len(domains))
	fmt.Printf("[*] عدد الخيوط المتزامنة: %d | المهلة: %ds\n\n", *concurrency, *timeoutSec)

	start := time.Now()

	jobs := make(chan string, len(domains))
	resultsCh := make(chan Result, len(domains))
	var wg sync.WaitGroup

	// إطلاق Worker Pool
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go worker(i, jobs, resultsCh, time.Duration(*timeoutSec)*time.Second, &wg)
	}

	// تغذية القناة بالنطاقات
	for _, d := range domains {
		jobs <- d
	}
	close(jobs)

	// إغلاق قناة النتائج بعد انتهاء جميع الـ workers
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

	fmt.Printf("\n[*] اكتمل الفحص في %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("[*] النتيجة: %d فعّال / %d إجمالي\n", aliveCount, len(domains))

	if *outJSON != "" {
		if err := exportJSON(results, *outJSON); err != nil {
			fmt.Printf("خطأ في تصدير JSON: %v\n", err)
		} else {
			fmt.Printf("[+] تم حفظ النتائج في: %s\n", *outJSON)
		}
	}

	if *outCSV != "" {
		if err := exportCSV(results, *outCSV); err != nil {
			fmt.Printf("خطأ في تصدير CSV: %v\n", err)
		} else {
			fmt.Printf("[+] تم حفظ النتائج في: %s\n", *outCSV)
		}
	}
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
