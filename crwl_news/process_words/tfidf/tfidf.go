package main

import (
	"fmt"
	"log"
	"os"

	"math"

	"../../models"
	"../../utils"
	wp "../wordprocess"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/tfidf.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

	page := 1
	condition := true
	var words = make(map[string][]string)
	var document_count = 0

	// construct keyword and calculate tf
	for true {
		content_news := models.LogContentNews{}
		logContents := log_content.Find(bson.M{"status": 4}).Sort("-_id").Skip((page - 1) * utils.LIMIT).Limit(utils.LIMIT).Iter()
		condition = false
		for logContents.Next(&content_news) {
			document_count++
			condition = true
			clean_data := wp.WordProcess{OriginContent: content_news.Content}
			clean_data.CleanData()
			clean_data.TokenizersbySpace()
			clean_data.RemoveStopwords()
			words = generateTermsCounter(clean_data.DocumentTermsMatrix, words, content_news.Id.Hex())

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
		// terms := wp.SortByKeys(words)
		log.Println(words)
		fmt.Println("len words and len terms", len(words))

		//Wait 30 minute until get more data
		if condition == false {
			break
		} else {
			page++
		}
	}

	idf_collection := calculateIDF(words, document_count)
	page = 1

	// update weight on each document by calculate tf-idf
	for true {
		content_news := models.LogContentNews{}
		logContents := log_content.Find(bson.M{"status": 5}).Sort("-_id").Skip((page - 1) * utils.LIMIT).Limit(utils.LIMIT).Iter()
		condition = false
		for logContents.Next(&content_news) {
			condition = true
			keyWords := content_news.KeyWords
			for _, keyword := range keyWords {
				if _, ok := idf_collection[keyword.Name]; ok {
					keyword.Total = keyword.Total * idf_collection[keyword.Name]
				}
			}

			// update status = 7 for tf-idf
			er_update := log_content.Update(
				bson.M{"_id": content_news.Id},
				bson.M{"$set": bson.M{
					"status":   7,
					"keywords": keyWords}})
			if er_update != nil {
				log.Println(er_update.Error())
			}
		}

		//Wait 30 minute until get more data
		if condition == false {
			break
		} else {
			page++
		}
	}
}

func generateTermDocumentMatrix(processs_data *wp.WordProcess) map[string]int {
	document := map[string]int{}
	for word, _ := range processs_data.DocumentTermsMatrix {
		if _, ok := document[word]; ok {
			document[word] = document[word] + 1
		} else {
			document[word] = 1
		}
	}
	return document
}

func generateTermsCounter(document map[string]float32, vocabulary map[string][]string, document_id string) map[string][]string {
	for word, _ := range document {
		if _, ok := vocabulary[word]; ok {
			vocabulary[word] = append(vocabulary[word], document_id)
		} else {
			vocabulary[word] = []string{document_id}
		}
	}
	return vocabulary
}

// construct keyword with normalize tf calculate
func constructKeyWord(pair_list wp.PairList) []models.Keyword {
	var keyWords = []models.Keyword{}
	max_fr := float32(0)
	for index, pair := range pair_list {
		if index == 0 {
			max_fr = pair.Value
		}
		if max_fr != 0 {
			keyword := models.Keyword{Name: pair.Key, Total: pair.Value / max_fr}
			keyWords = append(keyWords, keyword)
		}

	}
	return keyWords
}

func calculateIDF(words map[string][]string, document_collection int) map[string]float32 {
	var idf = map[string]float32{}
	for word, documents := range words {
		document_fr := len(documents)
		word_idf := math.Log2(float64(document_collection / document_fr))
		idf[word] = float32(word_idf)
	}
	return idf
}
