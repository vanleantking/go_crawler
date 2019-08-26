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
	vnexLinksC := local_client.Client.Database("docbao").Collection("vnexpress_links")
	// click view more comment button
	viewMoreE, er := detailDriver.FindElement(
		selenium.ByCSSSelector,
		".view_more_coment")
	if er == nil {
		viewMoreE.Click()
		time.Sleep(5 * time.Second)
	}

	// pagination
	paginationNextE, er := detailDriver.FindElement(
		selenium.ByCSSSelector,
		"#pagination a.next")

	detailComments := make([]structs.DetailComment, 0)
	if er != nil {
		detailComments = getAllDetailCmts(detailDriver, linkCrwl)
	} else {
		for {
			// get comment item from current page
			detailCmtPagin := getAllDetailCmts(detailDriver, linkCrwl)
			detailComments = append(detailComments, detailCmtPagin...)

			// find element next pagination
			paginationNextE, er = detailDriver.FindElement(
				selenium.ByCSSSelector,
				"#pagination a.next")
			if er != nil {
				break
			}
			// click next page
			paginationNextE.Click()
			time.Sleep(500 * time.Millisecond)

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

func getAllDetailCmts(detailDriver selenium.WebDriver, linkCrwl LinkCrwl) []structs.DetailComment {
	vnexUsersC := local_client.Client.Database("docbao").Collection("vnexpress_users")
	var allComments = make(map[string]structs.DetailComment)

	commentItemE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".comment_item")
	if er != nil {
		log.Println("eror on get comment_item value, ", er.Error())
		return make([]structs.DetailComment, 0)
	}

	// get text full comment
	// insert list of comment and user info
	for _, cmtItem := range commentItemE {
		// click txt_view_more comment
		for {
			viewMoreRepE, er := cmtItem.FindElements(
				selenium.ByCSSSelector, ".view_all_reply")
			fmt.Println("Error txt_view_more, ", er, len(viewMoreRepE))
			if len(viewMoreRepE) == 0 {
				break
			}

			for _, viewMoreRep := range viewMoreRepE {
				// click view more
				err := viewMoreRep.Click()
				href, _ := viewMoreRep.Text()
				fmt.Println("txt view more click err, ", err, viewMoreRepE, href)
				time.Sleep(1000 * time.Millisecond)
			}
		}

		// click icon_show_full_comment
		for {
			viewFullCmtsE, er := cmtItem.FindElements(
				selenium.ByCSSSelector,
				".content_less .icon_show_full_comment")
			fmt.Println("Error icon_show_full_comment, ", er, len(viewFullCmtsE))

			for _, viewFullCmtE := range viewFullCmtsE {
				// click view more
				err := viewFullCmtE.Click()
				fmt.Println("Error icon show full comment, ", err)
				time.Sleep(1000 * time.Millisecond)
			}
			break
		}

		// check length sub_comment length class:
		// == 0 => ((full_content, user_status) | (content_more, user_status))
		// (> 0) => ((full_content, user_status) | (content_more, user_status)),
		// (sub_comment[sub_comment_item(full_content, user_status)])

		// check length .sub_comment
		subCmtE, _ := cmtItem.FindElements(
			selenium.ByCSSSelector, ".sub_comment_item")
		if len(subCmtE) == 0 {
			detailCmt, er := getDetailCmt(cmtItem, linkCrwl)
			if er != nil || detailCmt.Content == "" {
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[detailCmt.Content]; !ok {
				allComments[detailCmt.Content] = detailCmt
			}
		} else {
			// get first comment
			firstComment, er := getDetailCmt(cmtItem, linkCrwl)
			if er != nil || firstComment.Content == "" {
				continue
			}
			// check detail comment not exist in map comments <= add
			if _, ok := allComments[firstComment.Content]; !ok {
				allComments[firstComment.Content] = firstComment
			}

			// get list of reply comment on sub_comment detail
			for _, subCmtDetail := range subCmtE {
				replyComment, er := getDetailCmt(subCmtDetail, linkCrwl)
				if er != nil || replyComment.Content == "" {
					continue
				}
				// check detail comment not exist in map comments <= add
				if _, ok := allComments[replyComment.Content]; !ok {
					allComments[replyComment.Content] = replyComment
				}
			}
		}
	}

	detailComments := make([]structs.DetailComment, 0)
	for _, detail := range allComments {

		linkPattern := regexp.MustCompile(LinkRexp)
		// only insert user if can  match link profile pattern
		if linkPattern.MatchString(detail.ProfileLink) {
			ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
			count, er := vnexUsersC.Count(ctx,
				bson.M{"profile_link": detail.ProfileLink})

			if er != nil {
				log.Println("Error, can not count profile from vnexpress_users, ", er.Error())
				continue
			}

			// only insert if not exist
			if count == int64(0) {
				user := structs.VNExUser{
					ID:          primitive.NewObjectID(),
					ProfileLink: detail.ProfileLink,
					Status:      0}

				ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
				_, er := vnexUsersC.InsertOne(
					ctx,
					&user)
				if er != nil {
					log.Println("Error, can not insert user, ", user, er.Error())
				}
			}
		}
		detailComments = append(detailComments, detail)
	}
	return detailComments
}

func getDetailCmt(cmtItem selenium.WebElement,
	linkCrwl LinkCrwl) (structs.DetailComment, error) {

	detailComment := structs.DetailComment{}
	// get user_info from comment element
	userInfoE, er := cmtItem.FindElement(
		selenium.ByCSSSelector, ".nickname")
	if er != nil {
		log.Println("eror on get user profile link value, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	profileLink, er := userInfoE.GetAttribute("href")
	if er != nil {
		log.Println("Error, can not get profile user, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	userName, er := userInfoE.Text()
	if er != nil {
		log.Println("Error, can not get user name, ", linkCrwl.Link, er.Error())
		return detailComment, er
	}

	detailComment.ProfileLink = profileLink
	detailComment.UserName = userName

	// click view full_content comment
	var fullCmtText = ""
	cmtsE, er := cmtItem.FindElement(
		selenium.ByCSSSelector, "p.full_content")
	if er == nil {
		fullCmtText, er = cmtsE.Text()
		if er != nil {
			log.Println("eror on get full_content comment value, ", linkCrwl, er.Error())
			return detailComment, er
		}
	} else {
		// click view content_more
		cmtmoresE, er := cmtItem.FindElement(
			selenium.ByCSSSelector,
			".content_more")
		if er != nil {
			log.Println("eror on get content_more view value, ", er.Error())
			return detailComment, er
		}

		fullCmtText, er = cmtmoresE.Text()
		if er != nil {
			return detailComment, er
		}
	}
	if fullCmtText == "" {
		return detailComment, er
	}
	detailComment.Content = fullCmtText
	return detailComment, nil
}
