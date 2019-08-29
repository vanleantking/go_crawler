package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"sync"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"regexp"

	"../paging"
	"../settings"
	"../utils"
	structs "./utils"
	"github.com/tebeka/selenium"
	"gopkg.in/mgo.v2/bson"
)

var (
	ipInfoClient *utils.ClientMGO
)

const (
	RegexUserID = `[0-9]+`
)

func main() {
	var wg sync.WaitGroup
	var er error

	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/scan_users.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)
	//test case
	log.Println("check to make sure it works")

	er, ipInfoClient = utils.ConnectMongoDB(utils.MongoDBInfo["ip_info"])
	if er != nil {
		panic(er.Error())
	}
	defer ipInfoClient.CancelFunc()
	defer ipInfoClient.Client.Disconnect(ipInfoClient.Ctx)

	scanUsers(&wg)
	scanLinks(&wg)
	wg.Wait() // wait until all thread has complete
	fmt.Println("Success.....................")

}

func scanUsers(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		vnexUsersC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_users")

		sortDesc := bson.M{"_id": -1}
		defaultCondition := bson.M{"status": 0}
		projectionFind := bson.M{
			"_id":          1,
			"profile_link": 1}
		paging := paging.NewPaging(
			vnexUsersC,
			defaultCondition,
			sortDesc, int64(1000))

		lastId, er := paging.GetMaxKey(
			projectionFind,
			sortDesc,
			defaultCondition)
		if er != nil {
			log.Println("Error, ", er.Error())
		}
		newId := lastId

		for true {
			// retry connect 3 times if error exist
			for iter := 0; iter < 3; iter++ {
				newId, er = paging.Paginage(bson.M{"$lte": lastId}, 30*time.Second)
				if er != nil {
					fmt.Println("scan users vnexpress_users Have no data to sync, please wait ", er.Error())
				} else {
					break
				}
			}
			if er != nil {
				fmt.Println("Error after 3 times retry, program exit", er.Error())
				break
			}
			if len(paging.Results) == 0 {
				fmt.Println("scan users vnexpress_users Have no data to sync, please wait")
				// wait 30 minute before next query
				time.Sleep(5 * time.Minute)
				// update last id
				lastId, er = paging.GetMaxKey(
					projectionFind,
					sortDesc,
					defaultCondition)
			} else {
				fmt.Println("len result, ", len(paging.Results))
				for _, result := range paging.Results {
					var err error
					var profileLink structs.ProfileUser

					bsonBytes, err := bson.Marshal(result)

					if err != nil {
						log.Println("Error: can not get decode go_cookies bson bytes", err.Error())
						continue
					}
					err = bson.Unmarshal(bsonBytes, &profileLink)
					if err != nil {
						log.Println("Error: can not get decode go_cookies models", err.Error())
						continue
					}

					updateLinksUserProfile(profileLink)
				}
				lastId = newId
			}
		}
	}()
}

func updateLinksUserProfile(profileLink structs.ProfileUser) {
	var webDriver selenium.WebDriver
	var er error
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()

	er = webDriver.Get(profileLink.ProfileLink)
	if er != nil {
		panic(er.Error())
	}
	vnexLinksC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_links")
	vnexUsersC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_users")

	// click load_more_comment button
	for {
		loadMoreContentsE, _ := webDriver.FindElements(
			selenium.ByCSSSelector,
			".xemthem a#load_more_comment")
		if len(loadMoreContentsE) == 0 {
			break
		}

		for _, loadMoreButton := range loadMoreContentsE {
			loadMoreButton.Click()
			time.Sleep(1500 * time.Microsecond)
		}
	}

	// get list of links
	articles, er := webDriver.FindElements(selenium.ByCSSSelector, ".list_activity .item_active")
	if er != nil {
		log.Println("eror on get articles value, ", er.Error())
		return
	}

	// insert all links to vnexpress_links from user comment
	for _, article := range articles {
		linkE, err := article.FindElement(selenium.ByCSSSelector, ".content_active p.title_article_com a")
		if err != nil {
			continue
		}

		link, err := linkE.Text()
		if err != nil {
			continue
		}

		href, err := linkE.GetAttribute("href")
		if err != nil {
			continue
		}
		urlParse, _ := url.Parse(href)
		originalLink := urlParse.Scheme + "://" + urlParse.Host + urlParse.Path

		ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
		count, er := vnexLinksC.Count(ctx,
			bson.M{"link": originalLink})

		if er != nil {
			log.Println("Error, can not count vnexpress_links, ", er.Error())
			continue
		}

		// only insert link if not exist
		if count == int64(0) {
			ctx, _ = context.WithTimeout(context.Background(), 5*time.Second)
			linkCrwl := structs.LinkCrawler{
				ID:     primitive.NewObjectID(),
				Link:   originalLink,
				Title:  link,
				Status: 0}
			_, er := vnexLinksC.InsertOne(ctx, &linkCrwl)
			if er != nil {
				log.Println("Error, can not insert vnexpress_links, ", er.Error())
				continue
			}
		}
	}

	userIDPattern := regexp.MustCompile(RegexUserID)
	matches := userIDPattern.FindAllString(profileLink.ProfileLink, -1)
	userNameE, er := webDriver.FindElement(selenium.ByCSSSelector, ".info_user_myvne p.name_myvne")
	userName := ""
	joinDate := ""
	if er == nil {
		userName, _ = userNameE.Text()
	}

	joinDateE, er := webDriver.FindElement(selenium.ByCSSSelector, ".info_user_myvne .join_date p span")
	if er == nil {
		joinDate, _ = joinDateE.Text()
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	_, er = vnexUsersC.UpdateOne(
		ctx,
		bson.M{"_id": profileLink.ID},
		bson.M{"$set": bson.M{
			"user_id":     matches[0],
			"user_name":   userName,
			"joined_date": joinDate,
			"status":      3}})
	if er != nil {
		log.Println("Error, can not insert vnexpress_users, ", er.Error())
	}
}

func scanLinks(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		vnexLinksC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_links")

		sortDesc := bson.M{"_id": -1}
		defaultCondition := bson.M{"status": 0}
		projectionFind := bson.M{
			"_id":  1,
			"link": 1}
		paging := paging.NewPaging(
			vnexLinksC,
			defaultCondition,
			sortDesc, int64(1000))

		lastId, er := paging.GetMaxKey(
			projectionFind,
			sortDesc,
			defaultCondition)
		if er != nil {
			log.Println("Error, ", er.Error())
		}
		newId := lastId

		for true {
			// retry connect 3 times if error exist
			for iter := 0; iter < 3; iter++ {
				newId, er = paging.Paginage(bson.M{"$lte": lastId}, 30*time.Second)
				if er != nil {
					fmt.Println("scan links vnexpress_links Have no data to sync, please wait ", er.Error())
				} else {
					break
				}
			}
			if er != nil {
				fmt.Println("Error after 3 times retry, program exit", er.Error())
				break
			}
			if len(paging.Results) == 0 {
				fmt.Println("scan links vnexpress_links Have no data to sync, please wait")
				// wait 30 minute before next query
				time.Sleep(5 * time.Minute)
				// update last id
				lastId, er = paging.GetMaxKey(
					projectionFind,
					sortDesc,
					defaultCondition)
			} else {
				fmt.Println("len result, ", len(paging.Results))
				for _, result := range paging.Results {
					var err error
					var profileLink structs.Link

					bsonBytes, err := bson.Marshal(result)

					if err != nil {
						log.Println("Error: can not get decode go_cookies bson bytes", err.Error())
						continue
					}
					err = bson.Unmarshal(bsonBytes, &profileLink)
					if err != nil {
						log.Println("Error: can not get decode go_cookies models", err.Error())
						continue
					}
					initProfileRequest(profileLink)
				}
				lastId = newId
			}
		}
	}()
}

func initProfileRequest(profileLink structs.Link) {
	var webDriver selenium.WebDriver
	var er error
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()
	vnexLinksC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_links")

	// client initial request on original url
	er = webDriver.Get(profileLink.Link)
	if er != nil {
		log.Println("Error on request link, ", er.Error(), profileLink.Link)
		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		_, er := vnexLinksC.UpdateOne(ctx,
			bson.M{"_id": profileLink.ID},
			bson.M{"$set": bson.M{
				"status": 2}})
		if er != nil {
			log.Println("Error, can not update vnexpress_links, ", er.Error())
		}
		return
	}
	time.Sleep(15 * time.Second)
	totalComments := 0

	totalCmtE, er := webDriver.FindElement(selenium.ByCSSSelector, "#total_comment")
	if er != nil {
		fmt.Println("Error, can not get total comment, ", er.Error())
	} else {
		totalCmts, _ := totalCmtE.Text()
		totalComments, _ = strconv.Atoi(totalCmts)
	}

	fmt.Println("--------------link crawl info receiver channel--------, ", profileLink)

	vnexUsersC := ipInfoClient.Client.Database("dmp_data").Collection("vnexpress_users")
	// click view more comment button
	viewMoreE, er := webDriver.FindElement(
		selenium.ByCSSSelector,
		".view_more_coment")
	if er == nil {
		viewMoreE.Click()
		time.Sleep(5 * time.Second)
	}

	// pagination
	_, er = webDriver.FindElement(
		selenium.ByCSSSelector,
		"#pagination a.next")

	var detailComments = []structs.DetailComment{}
	linkCrwl := structs.LinkCrwl{
		Link: profileLink.Link}
	if er != nil {
		detailComments = structs.GetAllDetailCmts(webDriver, linkCrwl, vnexUsersC)
	} else {
		countE := 0
		for {
			// get comment item from current page
			cmtPaginate := structs.GetAllDetailCmts(webDriver, linkCrwl, vnexUsersC)
			detailComments = append(detailComments, cmtPaginate...)

			// find element next pagination
			paginationsNextE, er := webDriver.FindElements(
				selenium.ByCSSSelector,
				"#pagination a.next")
			fmt.Println("Pagination-------------------------", len(detailComments),
				len(paginationsNextE), er, profileLink.Link)
			if len(paginationsNextE) == 0 {
				break
			}

			for _, paginationNext := range paginationsNextE {
				// click next page
				er := paginationNext.Click()
				if er != nil {
					countE++
					break
				}
				time.Sleep(3000 * time.Millisecond)
			}
			if countE > 0 {
				break
			}
		}
	}

	// update crawler link
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	_, er = vnexLinksC.UpdateOne(ctx,
		bson.M{"_id": profileLink.ID},
		bson.M{"$set": bson.M{
			"created":       time.Now().Unix(),
			"comments":      detailComments,
			"total_comment": totalComments,
			"status":        3}})
	if er != nil {
		log.Println("Error, can not update link, ", profileLink, er.Error())
	}
}
