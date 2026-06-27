package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/robertkrimen/otto"
)

func main() {
	url := "http://push2.eastmoney.com/api/qt/clist/get?cb=data&pn=1&pz=100&po=1&np=1&fltt=2&invt=2&fid=f62&fs=m:90+t:2+f:!50&fields=f12,f14,f2,f3,f62,f20"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 Edg/119.0.0.0")
	req.Header.Set("Host", "push2.eastmoney.com")
	// simulate setEastMoneyKlineBrowserHeaders
	req.Header.Set("Referer", "https://quote.eastmoney.com")
	
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	
	vm := otto.New()
	vm.Run("function data(res){return res};")
	val, err := vm.Run(string(body))
	if err != nil {
		fmt.Println("Otto Error:", err)
		return
	}
	value, _ := val.Export()
	marshal, _ := json.Marshal(value)
	var resData map[string]interface{}
	json.Unmarshal(marshal, &resData)
	
	out, _ := json.MarshalIndent(resData, "", "  ")
	fmt.Println(string(out))
}
