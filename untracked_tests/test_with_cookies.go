package main

import (
	"crypto/tls"
	"fmt"
	"go-stock/backend/data"
	"go-stock/backend/db"
	"io"
	"net/http"
	"time"
)

func main() {
	db.Init("./data/stock.db")
	config := data.GetSettingConfig()
	fmt.Printf("BrowserPath: %s\n", config.BrowserPath)

	cookieHeader := data.EastMoneyCookieHeaderForPush2his(config)
	fmt.Printf("Fetched Cookie Header: %q\n", cookieHeader)

	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           nil, // Explicitly bypass proxy to force direct connection
		},
	}

	urlStr := "https://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20"

	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequest("GET", urlStr, nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:146.0) Gecko/20100101 Firefox/146.0")
		req.Header.Set("Referer", "https://quote.eastmoney.com/center/gridlist.html")
		if cookieHeader != "" {
			req.Header.Set("Cookie", cookieHeader)
		}

		start := time.Now()
		resp, err := client.Do(req)
		duration := time.Since(start)
		if err != nil {
			fmt.Printf("Request %d FAIL: Duration: %v, Error: %v\n", i, duration, err)
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("Request %d SUCCESS: Status: %d, Length: %d, Duration: %v, Body snippet: %s\n", i, resp.StatusCode, len(body), duration, string(body[:100]))
		time.Sleep(1 * time.Second)
	}
}
