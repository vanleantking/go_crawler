package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"../model"
	"../paging"
	"./utils"
	"gopkg.in/mgo.v2/bson"

	"../crawler"
)

var (
	local_client *utils.ClientMGO
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
	var er error
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["localhost"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)
	var wg sync.WaitGroup
	fetchURL(&wg)
	crawlURL(&wg)
	wg.Wait()
	log.Println("Success")
	fmt.Println("Success")

}

func fetchURL(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		new_collection := local_client.Client.Database("docbao").Collection("news")
		crwl := crawler.InitCrawler()
		for {
			links := crwl.FetchURL()
			for _, link := range links {
				ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
				count, er := new_collection.Count(
					ctx,
					bson.M{"url": link})
				if er != nil {
					log.Println("Can not count news, ", link, er.Error())
					continue
				}
				if count == 0 {
					created_int := currentTimeUnix()
					new := model.News{URL: link, CreatedInt: created_int.Unix(), Status: 1}
					new.Id = primitive.NewObjectID()
					ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
					_, er := new_collection.InsertOne(ctx, &new)
					if er != nil {
						log.Println("Error, can not insert link", link, er.Error())
						continue
					}
				}
			}

			// break 30 mins before next crawl
			time.Sleep(30 * time.Minute)
			crwl.NewClient()
		}
	}()
}

func crawlURL(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		default_condition := bson.M{"status": bson.M{"$in": []int{4, 1}}}
		projection := bson.M{"_id": 1}
		sortDesc := bson.M{"_id": 1}
		new_collection := local_client.Client.Database("docbao").Collection("news")

		paging := paging.NewPaging(
			new_collection,
			default_condition,
			sortDesc, int64(1000))

		lastId, er := paging.GetMaxKey(
			projection,
			sortDesc,
			default_condition)
		if er != nil {
			log.Println("Error, ", er.Error())
		}
		newId := lastId
		var cookies_chan = make(chan bson.M)
		var done = make(chan bool)
		crwl := crawler.InitCrawler()
		// go-routine for send data cho channels
		go func() {
			for {
				newId, er = paging.Paginage(bson.M{"$gte": lastId}, 30*time.Second)
				if er != nil {
					log.Println("Error on get data, ", er.Error())
					fmt.Println("Have no data to sync, please wait ", er.Error())
					break
				}
				if len(paging.Results) == 0 {
					fmt.Println("Have no data to sync, please wait")
					log.Println("Have no data to sync, please wait")
					// wait 30 minute before next query
					time.Sleep(30 * time.Minute)
					// update last id
					lastId, er = paging.GetMaxKey(
						projection,
						sortDesc,
						default_condition)
				} else {
					// Send the hits to the hits channel
					for _, result := range paging.Results {
						select {
						case cookies_chan <- result:
						}
					}
					lastId = newId
				}
			}
			done <- true
		}()
		// group of routine for get string content update on cookie_full_v2
		for i := 0; i < 20; i++ {
			go func() {
				count := 1
				for cookie_chan := range cookies_chan {
					count++

					// setting new client each request
					crwl.NewClient()
					var err error
					var news model.News
					bsonBytes, err := bson.Marshal(cookie_chan)

					if err != nil {
						log.Println("----------------Error-------------: can not get decode go_cookies bson bytes", err.Error())
						continue
					}
					err = bson.Unmarshal(bsonBytes, &news)
					if err != nil {
						log.Println("----------------Error-------------: can not get decode go_cookies models", err.Error())
						continue
					}
					result, er := crwl.CrawlerURL(news.URL)
					u, err := url.Parse(news.URL)
					if err != nil {
						log.Println("Error, can not get domain, ", err.Error())
					}
					domain := utils.GetDomainName(u.Hostname())
					if er != nil {
						log.Println("Error, can not crawl content, ", news.Id, er.Error())
						ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
						_, er := new_collection.UpdateOne(
							ctx,
							bson.M{"_id": news.Id},
							bson.M{"$set": bson.M{
								"status":      4,
								"domain":      domain,
								"date_time":   currentTimeUnix().Format("2006-01-02"),
								"updated_str": currentTimeUnix().Format("2006-01-02 15:04:05")}})
						if er != nil {
							log.Println("Error, can not update status news, ", news.Id, er.Error())
							continue
						}
						continue
					} else {
						ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
						_, er = new_collection.UpdateOne(
							ctx,
							bson.M{"_id": news.Id},
							bson.M{"$set": bson.M{
								"title":         result.Title,
								"content":       result.Content,
								"category_news": result.CategoryNews,
								"description":   result.Description,
								"keyword":       result.Keyword,
								"meta":          result.Meta,
								"publish_date":  result.PublishDate,
								"domain":        domain,
								"status":        2,
								"date_time":     currentTimeUnix().Format("2006-01-02"),
								"updated_str":   currentTimeUnix().Format("2006-01-02 15:04:05")}})
						if er != nil {
							log.Println("Error on get content, ", news.Id, er.Error())
							continue
						}
						log.Println("success")
					}
				}
			}()
		}
		<-done
	}()

}

func currentTimeUnix() time.Time {
	//init the loc
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	//set timezone,
	return time.Now().In(loc)
}
