package main

import (
	"fmt"
	"log"
	"os"

	"./crawler"
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/vietgiaitri.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	//test case
	log.Println("check to make sure it works")

	crawler := &crawler.Crawler{}
	crawler.NewClient()
	log.Println("enter crawler")
	er := crawler.CrawlerURL("https://vnexpress.net/phap-luat/cuu-bo-truong-truong-minh-tuan-bi-bat-3885283.html")
	log.Println("out crawler", er)
	_, _, _, _, keywords, metas := crawler.Getresult()
	fmt.Println(keywords, metas)
	log.Println("success")
}
