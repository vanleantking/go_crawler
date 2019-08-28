package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"gopkg.in/mgo.v2/bson"

	"context"
	"net/url"
	"strconv"

	structs "./utils"

	"../settings"
	"../utils"
	"github.com/tebeka/selenium"
)

var (
	ipInfoClient *utils.ClientMGO
)

func main() {

	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/update_proxy.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)
	//test case
	log.Println("check to make sure it works")

	var er error
	er, ipInfoClient = utils.ConnectMongoDB(utils.MongoDBInfo["ip_info"])
	if er != nil {
		panic(er.Error())
	}
	defer ipInfoClient.CancelFunc()
	defer ipInfoClient.Client.Disconnect(ipInfoClient.Ctx)

	vnexLinksC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_links")

	// initialize value for selenium
	var webDriver selenium.WebDriver
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()

	var linkChan = make(chan structs.LinkCrwl)
	var done = make(chan bool)

	go func() {
		for {
			// client initial request on original url
			er = webDriver.Get("https://vnexpress.net/")
			if er != nil {
				panic(er.Error())
			}

			// select type anonymous
			articles, er := webDriver.FindElements(selenium.ByCSSSelector, "article.list_news")
			if er != nil {
				log.Println("eror on get articles value, ", er.Error())
				return
			}

			for _, article := range articles {
				linkE, err := article.FindElement(selenium.ByCSSSelector, "h4.title_news a")
				if err != nil {
					continue
				}

				link, err := linkE.Text()
				if err != nil {
					continue
				}

				totalCmtE, err := article.FindElement(selenium.ByCSSSelector, "span.txt_num_comment")
				if err != nil {
					continue
				}
				cmt, err := totalCmtE.Text()
				if err != nil {
					continue
				}

				totalCom, err := strconv.Atoi(cmt)
				if err != nil {
					continue
				}

				href, err := linkE.GetAttribute("href")
				if err != nil {
					continue
				}
				linkCrwl := structs.LinkCrwl{
					Link:     href,
					Title:    link,
					TotalCmt: totalCom}
				fmt.Println("link crawl info sender channel, ", linkCrwl)
				if totalCom > 0 {
					linkChan <- linkCrwl
				}
			}

			time.Sleep(30 * time.Minute)
			_, er = webDriver.NewSession()
			if er != nil {
				log.Println("Can not initial new session, ", er.Error())
				continue
			}
		}
		done <- true
	}()

	for i := 0; i < 3; i++ {
		go func() {
			for linkCrwl := range linkChan {
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				urlParse, _ := url.Parse(linkCrwl.Link)
				originalLink := urlParse.Scheme + "://" + urlParse.Host + urlParse.Path

				// update original link
				linkCrwl.Link = originalLink

				count, er := vnexLinksC.Count(
					ctx,
					bson.M{"link": originalLink})

				if er != nil {
					log.Println("Error, can not count link, ", linkCrwl.Link, er.Error())
					continue
				}

				if count > 0 {
					continue
				}

				initRequest(linkCrwl)
			}
		}()
	}
	<-done
}

// insert new link and user vnexpress
// extract list users from each page_url
func initRequest(linkCrwl structs.LinkCrwl) {
	var detailDriver selenium.WebDriver
	var er error
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if detailDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer detailDriver.Quit()
	// client initial request on original url
	er = detailDriver.Get(linkCrwl.Link)
	if er != nil {
		log.Println("Error on request link, ", er.Error(), linkCrwl.Link)
		return
	}
	time.Sleep(10 * time.Second)

	fmt.Println("--------------link crawl info receiver channel--------, ", linkCrwl)
	vnexLinksC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_links")
	vnexUsersC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_users")
	// click view more comment button
	viewMoreE, er := detailDriver.FindElement(
		selenium.ByCSSSelector,
		".view_more_coment")
	if er == nil {
		viewMoreE.Click()
		time.Sleep(5 * time.Second)
	}

	// pagination
	_, er = detailDriver.FindElement(
		selenium.ByCSSSelector,
		"#pagination a.next")

	var detailComments = []structs.DetailComment{}
	if er != nil {
		detailComments = structs.GetAllDetailCmts(detailDriver, linkCrwl, vnexUsersC)
	} else {
		countE := 0
		for {
			// get comment item from current page
			cmtPaginate := structs.GetAllDetailCmts(detailDriver, linkCrwl, vnexUsersC)
			detailComments = append(detailComments, cmtPaginate...)

			// find element next pagination
			paginationsNextE, er := detailDriver.FindElements(
				selenium.ByCSSSelector,
				"#pagination a.next")
			fmt.Println("Pagination-------------------------", len(detailComments),
				len(paginationsNextE), er, linkCrwl.Link)
			if len(paginationsNextE) == 0 {
				break
			}

			for _, paginationNext := range paginationsNextE {
				// click next page
				er := paginationNext.Click()
				fmt.Println("Err paginationnnnnnnnnnnnnnnnn, ", er)
				if er != nil {
					countE++
					break
				}
				time.Sleep(2 * time.Second)
			}
			if countE > 0 {
				break
			}
		}
	}

	// insert crawler link
	linkCrw := structs.LinkCrawler{
		ID:           primitive.NewObjectID(),
		Link:         linkCrwl.Link,
		Comments:     detailComments,
		Created:      time.Now().Unix(),
		TotalComment: linkCrwl.TotalCmt,
		Title:        linkCrwl.Title,
		Status:       3}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, er = vnexLinksC.InsertOne(ctx, &linkCrw)
	if er != nil {
		log.Println("Error, can not insert link, ", linkCrw, er.Error())
	}
}
