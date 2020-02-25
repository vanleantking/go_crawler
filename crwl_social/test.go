package main

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"../crawler"
	"../utils"
)

var ()

const (
	REVIEW     = "review"
	PAGE_LIMIT = 50
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/roto.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)
	//test case
	log.Println("check to make sure it works")
	crawlURL()
	log.Println("Success")
	fmt.Println("Success")

}

func crawlURL() {
	// setting new client each request
	crwl := crawler.InitCrawler()
	var err error

	crawl_url := "https://tinhte.vn/threads/michelin-va-gm-thu-nghiem-lop-khong-hoi-cho-xe-du-lich-du-kien-thuong-mai-vao-nam-2024.2970179/"
	u, err := url.Parse(crawl_url)
	if err != nil {
		log.Println("Error, can not get domain, ", err.Error())
	}
	domain := utils.GetDomainName(u.Hostname())
	ws := crwl.WS[domain]
	var cmts = map[string]bool{}

	// initialized last state for fisrst run
	domain_state := crawler.StateDomain{
		CurrentPage: 0,
		ErrCode:     0,
		Status:      true}

	if ws.PaginateRegex != "" {
		for true {
			if domain_state.ErrCode < 400 && domain_state.Status {
				domain_state.CurrentPage += 1
			} else {
				// request to limit page of thread <= server return 404
				if domain_state.ErrCode == 404 {
					break
				}
			}

			// update crawl_url when page > 1
			pcrawl_url := crawl_url
			if domain_state.CurrentPage > 1 {
				pcrawl_url = fmt.Sprintf(crawl_url+ws.PaginateRegex, domain_state.CurrentPage)
			}

			result, err, res_code := crwl.GetResultCrwl(pcrawl_url)

			// the last comment already exist in map string <= repeat crawl
			// <= server exist to limit page thread but still return 200
			len_reviews := len(result.Reviews)
			if len_reviews == 0 {
				continue
			}

			last_review := result.Reviews[len_reviews-1]
			if _, ok := cmts[last_review]; ok {
				break
			}

			for _, review := range result.Reviews {
				cmts[review] = true
			}
			fmt.Println("INfo, ", err, res_code, domain_state, pcrawl_url)

			if err != nil {
				domain_state.Status = false
				domain_state.ErrCode = res_code
			} else {
				domain_state.Status = true
				domain_state.ErrCode = 0
			}
		}
	}

	fmt.Println(cmts)
}
