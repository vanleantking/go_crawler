package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"./crawler"
	"./model"
	"./paging"
	"./utils"
	"gopkg.in/mgo.v2/bson"
)

var (
	local_client *utils.ClientMGO
)

func main() {
	var er error
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["localhost"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)

	default_condition := bson.M{"status": 2, "content": ""}
	projection := bson.M{"_id": 1}
	sortDesc := bson.M{"_id": 1}
	new_collection := local_client.Client.Database("docbao").Collection("news")
	crwl := crawler.InitCrawler()

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
				var news model.News
				bsonBytes, err := bson.Marshal(result)

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
			lastId = newId
		}
	}
}

func currentTimeUnix() time.Time {
	//init the loc
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")

	//set timezone,
	return time.Now().In(loc)
}
