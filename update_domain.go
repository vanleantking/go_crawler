package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

	default_condition := bson.M{"status": 2, "date_time": nil}
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
				date_time := strings.Split(news.UpdatedStr, " ")[0]
				ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
				_, _ = new_collection.UpdateOne(
					ctx,
					bson.M{"_id": news.Id},
					bson.M{"$set": bson.M{
						"date_time": date_time}})
			}
			lastId = newId
		}
	}
}
