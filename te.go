package main

import (
	"fmt"

	"./crawler"
)

func main() {
	crawler := &crawler.Crawler{}
	crawler.NewClient()
	crawler.CrawlerURL("https://vnexpress.net/hoi-nghi-thuong-dinh-my-trieu/trump-loai-tru-kha-nang-giam-quan-my-o-han-quoc-3885129.html")
	fmt.Println(crawler.CrResult)
}
