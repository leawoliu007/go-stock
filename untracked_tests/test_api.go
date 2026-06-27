package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

func testGet(urlStr string, useHostHeader bool) {
	fmt.Printf("Testing URL: %s (useHostHeader: %v)\n", urlStr, useHostHeader)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		fmt.Printf("NewRequest error: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	if useHostHeader {
		req.Header.Set("Host", "push2.eastmoney.com")
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Do error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Read body error: %v\n", err)
		return
	}

	fmt.Printf("Status: %d, Body length: %d\n", resp.StatusCode, len(body))
	if len(body) > 0 {
		preview := string(body)
		if len(preview) > 200 {
			preview = preview[:200]
		}
		fmt.Printf("Body preview: %s\n", preview)
	}
}

func main() {
	// Test HTTPS with Host header
	testGet("https://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20", true)
	fmt.Println()

	// Test HTTPS without Host header
	testGet("https://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20", false)
	fmt.Println()

	// Test HTTP without Host header
	testGet("http://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20", false)
	fmt.Println()
}
