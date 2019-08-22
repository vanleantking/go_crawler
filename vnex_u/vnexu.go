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
	"regexp"
	"strconv"

	structs "./utils"

	"../settings"
	"../utils"
	"github.com/tebeka/selenium"
)

var (
	local_client *utils.ClientMGO
)

type LinkCrwl struct {
	Link     string
	TotalCmt int
	Title    string
}

const (
	LinkRexp = `\b(https?)`
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
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["docbao"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)

	vnexLinksC := local_client.Client.Database("docbao").Collection("vnexpress_links")

	// initialize value for selenium
	var webDriver selenium.WebDriver
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()

	var linkChan = make(chan LinkCrwl)
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
				linkCrwl := LinkCrwl{
					Link:     href,
					Title:    link,
					TotalCmt: totalCom}
				fmt.Println("link crawl info sender channel, ", linkCrwl)
				if totalCom > 0 {
					linkChan <- linkCrwl
				}
			}

			time.Sleep(10 * time.Minute)
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
func initRequest(linkCrwl LinkCrwl) {
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
	fmt.Println("--------------link crawl info receiver channel--------, ", linkCrwl)
	// click view more comment button
	viewMoreE, er := detailDriver.FindElement(
		selenium.ByCSSSelector,
		".view_more_coment")
	if er == nil {
		viewMoreE.Click()
		time.Sleep(5 * time.Second)
	}

	// pagination
	paginationsE, er := detailDriver.FindElements(
		selenium.ByCSSSelector,
		"#pagination.pagination a")
	fmt.Println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz, ",
		paginationsE,
		er,
		len(paginationsE),
		linkCrwl.Link)
	if er != nil {
		log.Println("Error, can not get pagination, ")
	} else {
		if len(paginationsE) == 0 {
			getDetailCmt(detailDriver, linkCrwl)
		} else {
			maxLengthE := paginationsE[len(paginationsE)-2]
			turnRight := paginationsE[len(paginationsE)-1]
			trt, _ := turnRight.Text()
			maxLengthText, _ := maxLengthE.Text()
			maxLength, _ := strconv.Atoi(maxLengthText)
			fmt.Println("rrrrrrrrrrrrrrrrrrrrrr, ",
				maxLengthText,
				trt,
				len(paginationsE),
				len(paginationsE)-2,
				linkCrwl.Link)
			for i := 0; i < maxLength-1; i++ {
				turnRight.Click()
				time.Sleep(100 * time.Millisecond)
				getDetailCmt(detailDriver, linkCrwl)
			}
		}
	}
}

func getDetailCmt(detailDriver selenium.WebDriver, linkCrwl LinkCrwl) {
	vnexUsersC := local_client.Client.Database("docbao").Collection("vnexpress_users")
	vnexLinksC := local_client.Client.Database("docbao").Collection("vnexpress_links")
	var allComments = make(map[string]structs.DetailComment)
	// click view full comment
	viewFullCmtsE, er := detailDriver.FindElements(
		selenium.ByCSSSelector,
		".comment_item .content_less .icon_show_full_comment")
	if er == nil {
		// click view more
		for _, viewFull := range viewFullCmtsE {
			viewFull.Click()
			time.Sleep(200 * time.Millisecond)
		}
	}

	// click view more reply comment
	viewMoreRepE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".txt_view_more")
	if er == nil {
		// click view more
		for _, viewMore := range viewMoreRepE {
			viewMore.Click()
			time.Sleep(200 * time.Millisecond)
		}
	}

	commentItemE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".comment_item")
	if er != nil {
		log.Println("eror on get comment_item value, ", er.Error())
		return
	}

	// get text full comment
	// insert list of comment and user info
	for _, cmtItem := range commentItemE {
		// get user_info from comment element
		userInfoE, er := cmtItem.FindElement(
			selenium.ByCSSSelector, ".user_status .nickname")
		if er != nil {
			log.Println("eror on get user profile link value, ", linkCrwl.Link, er.Error())
			return
		}

		profileLink, er := userInfoE.GetAttribute("href")
		if er != nil {
			log.Println("Error, can not get profile user, ", linkCrwl.Link, er.Error())
			continue
		}

		userName, er := userInfoE.Text()
		if er != nil {
			log.Println("Error, can not get user name, ", linkCrwl.Link, er.Error())
			continue
		}

		detailComment := structs.DetailComment{
			ProfileLink: profileLink,
			UserName:    userName}

		// click view full_content comment
		var fullCmtText = ""
		cmtsE, er := cmtItem.FindElement(
			selenium.ByCSSSelector, "p.full_content")
		if er == nil {
			fullCmtText, er = cmtsE.Text()
			if er != nil {
				log.Println("eror on get full_content comment value, ", linkCrwl, er.Error())
				continue
			}
		} else {
			// click view content_more
			cmtmoresE, er := detailDriver.FindElement(
				selenium.ByCSSSelector,
				".comment_item .content_more")
			if er != nil {
				log.Println("eror on get content_more view value, ", er.Error())
				return
			}

			fullCmtText, er = cmtmoresE.Text()
			if er != nil {
				continue
			}
		}
		if fullCmtText == "" {
			continue
		}
		detailComment.Content = fullCmtText

		// check detail comment not exist in map comments <= add
		if _, ok := allComments[fullCmtText]; !ok {
			allComments[fullCmtText] = detailComment
		}
	}

	detailComment := make([]structs.DetailComment, 0)
	for _, detail := range allComments {

		linkPattern := regexp.MustCompile(LinkRexp)
		// only insert user if can  match link profile pattern
		if linkPattern.MatchString(detail.ProfileLink) {
			user := structs.VNExUser{
				ID:          primitive.NewObjectID(),
				ProfileLink: detail.ProfileLink,
				Status:      0}

			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			_, er := vnexUsersC.InsertOne(
				ctx,
				&user)
			if er != nil {
				log.Println("Error, can not insert user, ", user, er.Error())
			}
		}
		detailComment = append(detailComment, detail)
	}

	// insert crawler link
	linkCrw := structs.LinkCrawler{
		ID:           primitive.NewObjectID(),
		Link:         linkCrwl.Link,
		Comments:     detailComment,
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