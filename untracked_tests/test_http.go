package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	urls := []string{
		"https://quote.eastmoney.com/",
		"https://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20",
	}

	for _, u := range urls {
		fmt.Printf("Fetching: %s\n", u)
		req, _ := http.NewRequest("GET", u, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		
		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)
		if err != nil {
			fmt.Printf("FAIL: %s (Duration: %v, Error: %v)\n", u, duration, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("SUCCESS: %s (Status: %d, Length: %d, Duration: %v, Body snippet: %s)\n", u, resp.StatusCode, len(body), duration, string(body[:100]))
	}
}
