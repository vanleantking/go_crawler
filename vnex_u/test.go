package main

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"regexp"

	structs "./utils"

	"../settings"
	"../utils"
	"github.com/tebeka/selenium"
)

var (
// local_client *utils.ClientMGO
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
	f, err := os.OpenFile("./log/test.update_proxy.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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
	// er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["docbao"])
	// if er != nil {
	// 	panic(er.Error())
	// }
	// defer local_client.CancelFunc()
	// defer local_client.Client.Disconnect(local_client.Ctx)

	// vnexLinksC := local_client.Client.Database("docbao").Collection("vnexpress_links")

	// initialize value for selenium
	var webDriver selenium.WebDriver
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()
	linkCrwl := "https://news.zing.vn/"
	// client initial request on original url
	er = webDriver.Get(linkCrwl)
	if er != nil {
		panic(er.Error())
	}
	time.Sleep(5 * time.Second)
	elm, er := webDriver.FindElement(selenium.ByCSSSelector, "html")
	if er != nil {
		panic(er.Error())
	}
	// time.Sleep(5 * time.Second)

	size, er := elm.Size()
	if er != nil {
		panic(er.Error())
	}
	fmt.Println("size, ", size.Width, size.Height, er)
	zzz, _ := elm.IsDisplayed()
	fmt.Println("isdisplayed, ", zzz)

	// point, er := elm.LocationInView()
	// if er != nil {
	// 	panic(er.Error())
	// }
	// fmt.Println("point, ", point.X, point.Y, er)
	er = elm.MoveTo(size.Width, size.Height)
	if er != nil {
		panic(er.Error())
	}
	time.Sleep(5 * time.Second)
	bb, er := elm.Screenshot(true)
	if er != nil {
		fmt.Println("err -", er)
		panic(er.Error())
	}

	img, _, _ := image.Decode(bytes.NewReader(bb))
	out, err := os.Create("./img/sreenshot" + "name" + ".png")
	if err != nil {
		fmt.Println("err -", err)
		panic(err.Error())
	}

	err = png.Encode(out, img)
	if err != nil {
		panic(err.Error())
	}

	// zinitRequest(linkCrwl)
}

func SaveImage(foto []byte, name string) error {
	img, _, _ := image.Decode(bytes.NewReader(foto))
	out, err := os.Create("./img/sreenshot" + name + ".png")
	if err != nil {
		fmt.Println("err -", err)
		return err
	}

	err = png.Encode(out, img)
	if err != nil {
		return err
	}
	return nil
}

// insert new link and user vnexpress
// extract list users from each page_url
func zinitRequest(linkCrwl string) {
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
	er = detailDriver.Get(linkCrwl)
	if er != nil {
		log.Println("Error on request link, ", er.Error(), linkCrwl)
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
	paginationNextE, er := detailDriver.FindElement(
		selenium.ByCSSSelector,
		"#pagination a.next")
	fmt.Println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz, ",
		paginationNextE,
		er,
		linkCrwl)
	if er != nil {
		log.Println("Error, can not get pagination, ")
		getAllDetailCmts(detailDriver, linkCrwl)
	} else {
		for {
			// get comment item from current page
			getAllDetailCmts(detailDriver, linkCrwl)

			// find element next pagination
			paginationsNextE, er := detailDriver.FindElements(
				selenium.ByCSSSelector,
				"#pagination a.next")
			log.Println("Pagination-------------------------", len(paginationsNextE), er)
			if len(paginationsNextE) == 0 {
				break
			}

			for _, paginationNext := range paginationsNextE {
				// click next page
				paginationNext.Click()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}
}

func getAllDetailCmts(detailDriver selenium.WebDriver, linkCrwl string) {
	var allComments = make(map[string]structs.DetailComment)

	commentItemE, er := detailDriver.FindElements(
		selenium.ByCSSSelector, ".comment_item")
	if er != nil {
		log.Println("eror on get comment_item value, ", er.Error())
		return
	}

	// get text full comment
	// insert list of comment and user info
	for _, cmtItem := range commentItemE {
		// click view more reply comment
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

		// click view full comment
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
		subCmtE, er := cmtItem.FindElements(
			selenium.ByCSSSelector, ".sub_comment_item")
		log.Println("Error on get sub comment item, ", er, len(subCmtE))
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
			user := structs.VNExUser{
				ID:          primitive.NewObjectID(),
				ProfileLink: detail.ProfileLink,
				Status:      0}

			log.Println("User info, ", user)
		}
		detailComments = append(detailComments, detail)
	}

	// insert crawler link
	log.Println("detail comments, ", detailComments)

}

func getDetailCmt(cmtItem selenium.WebElement,
	linkCrwl string) (structs.DetailComment, error) {

	clsCmtItem, _ := cmtItem.Text()

	detailComment := structs.DetailComment{}
	// get user_info from comment element
	userInfoE, er := cmtItem.FindElement(
		selenium.ByCSSSelector, ".nickname")
	if er != nil {
		log.Println("eror on get user profile link value, ", linkCrwl, er.Error(), "class item,", clsCmtItem)
		return detailComment, er
	}

	profileLink, er := userInfoE.GetAttribute("href")
	if er != nil {
		log.Println("Error, can not get profile user, ", linkCrwl, er.Error())
		return detailComment, er
	}

	userName, er := userInfoE.Text()
	if er != nil {
		log.Println("Error, can not get user name, ", linkCrwl, er.Error())
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
