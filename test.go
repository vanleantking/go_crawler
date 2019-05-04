package main

import (
	"fmt"

	"./crawler"
)

func main() {
	crawler := &crawler.Crawler{}
	crawler.NewClient()

	er := crawler.CrawlerURL("http://cafebiz.vn/phat-hien-cocaine-trong-tom-nuoc-ngot-tai-anh-20190504133747787.chn")
	if er != nil {
		fmt.Println("Error, can not crawl content, ", er.Error())
	} else {
		result := crawler.Getresult()
		fmt.Println(result.Content)
	}
}
