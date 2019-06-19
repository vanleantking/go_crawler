package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"../crawler"
	"../model"
	"../paging"
	"../settings"
	"../utils"
	"gopkg.in/mgo.v2/bson"
)

var (
	local_client *utils.ClientMGO
)
var (
	RegexpProxy = `[a-z0-9\\.]+`
)

func main() {
	var er error
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["docbao"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)

	var wg sync.WaitGroup

	crawlURL(&wg)
	wg.Wait()
}

func crawlURL(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		default_condition := bson.M{"category_type": "review", "status": bson.M{"$in": []int{4, 1}}}
		projection := bson.M{"_id": 1}
		sortDesc := bson.M{"_id": 1}
		new_collection := local_client.Client.Database("docbao").Collection("domain_xe")
		proxy_collection := local_client.Client.Database("docbao").Collection("proxy")

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
						_, err := new_collection.UpdateOne(
							ctx,
							bson.M{"_id": news.Id},
							bson.M{"$set": bson.M{
								"status":      4,
								"domain":      domain,
								"date_time":   currentTimeUnix().Format("2006-01-02"),
								"updated_str": currentTimeUnix().Format("2006-01-02 15:04:05")}})
						if err != nil {
							log.Println("Error, can not update status news, ", news.Id, err.Error())
							continue
						}

						// update status proxy on init request failed
						proxy_str := strings.Replace(er.Error(), settings.ErrProxyPrefix, "", -1)
						re := regexp.MustCompile(RegexpProxy)

						proxy_pieces := re.FindAllString(strings.TrimSpace(proxy_str), -1)
						ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
						_, er = proxy_collection.UpdateOne(
							ctx,
							bson.M{
								"proxy_ip": strings.TrimSpace(proxy_pieces[1]),
								"port":     strings.TrimSpace(proxy_pieces[2]),
								"schema":   strings.TrimSpace(proxy_pieces[0])},
							bson.M{"$set": bson.M{"status": false}})
						if er != nil {
							log.Println("Can not update status proxy, ", er.Error())
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
								"publish_date":  strings.TrimSpace(result.PublishDate),
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
