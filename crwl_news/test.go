package main

import (
	"fmt"

	"../crawler"
)

func main() {
	crwl := crawler.InitCrawler()

	links := crwl.FetchURLs("https://vnexpress.net/kinh-doanh")
	fmt.Println(links)
}
