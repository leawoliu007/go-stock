package main

import (
	"fmt"
	"go-stock/backend/db"
	"go-stock/backend/data"
)

func main() {
	db.Init("./data/stock.db")
	settings := &data.Settings{}
	db.Dao.Model(&data.Settings{}).First(settings)
	fmt.Printf("HttpProxyEnabled: %t\n", settings.HttpProxyEnabled)
	fmt.Printf("HttpProxy: %s\n", settings.HttpProxy)
	fmt.Printf("BrowserPath: %s\n", settings.BrowserPath)
}
