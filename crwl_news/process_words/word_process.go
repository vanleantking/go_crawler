package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"../models"
	"../utils"
	wp "./wordprocess"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	var wg sync.WaitGroup
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/content_news.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	//test case
	log.Println("check to make sure it works")

	// connect local to test
	local_session, err := mgo.Dial(utils.LOCALHOST)
	if err != nil {
		panic(err.Error())
	}
	defer local_session.Close()
	log_content := local_session.DB("dmp_cookies_only_v2").C("elog_content_news")

	range0_49 := bson.M{"$lte": 49, "$gte": 0}
	range50_99 := bson.M{"$lte": 99, "$gte": 50}

	wordprocess(log_content, range0_49, &wg)
	wordprocess(log_content, range50_99, &wg)

	wg.Wait() // wait until all thread has complete

	fmt.Println("Success.....................")

	// }
}

func wordprocess(log_content *mgo.Collection, rand_range bson.M, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		page := 1
		condition := true

		for true {
			content_news := models.LogContentNews{}
			logContents := log_content.Find(bson.M{"status": 4, "rand_number": rand_range}).Sort("-_id").Skip((page - 1) * utils.LIMIT).Limit(utils.LIMIT).Iter()
			condition = false
			for logContents.Next(&content_news) {
				condition = true
				clean_data := wp.WordProcess{OriginContent: content_news.Content}
				clean_data.CleanData()
				clean_data.TokenizersbySpace()
				clean_data.RemoveStopwords()

				// represent document by term-document matrix
				pl := wp.SortByFrequencies(clean_data.DocumentTermsMatrix)
				keyWords := constructKeyWord(pl)
				er_update := log_content.Update(
					bson.M{"_id": content_news.Id},
					bson.M{"$set": bson.M{
						"status":   5,
						"keywords": keyWords}})
				if er_update != nil {
					log.Println(er_update.Error())
				}
				log.Println(pl)
			}

			//Wait 30 minute until get more data
			if condition == false {
				break
			} else {
				page++
			}
		}
	}()
}

func constructKeyWord(pair_list wp.PairList) []models.Keyword {
	var keyWords = []models.Keyword{}
	for _, pair := range pair_list {
		keyword := models.Keyword{Name: pair.Key, Total: pair.Value}
		keyWords = append(keyWords, keyword)
	}
	return keyWords
}
