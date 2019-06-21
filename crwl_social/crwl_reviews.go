package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	"../model"
	"../paging"
	"../utils"
	"gopkg.in/mgo.v2/bson"

	"regexp"

	"../crawler"
	"../settings"
)

var (
	local_client *utils.ClientMGO
	RegexpProxy  = `[a-z0-9\\.]+`
)

const (
	REVIEW     = "review"
	PAGE_LIMIT = 50
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/oto.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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
	var wg sync.WaitGroup

	last_state := &crawler.LastRun{}
	fetchReviewsURL(&wg, last_state)
	crawlURL(&wg)
	wg.Wait()
	log.Println("Success")
	fmt.Println("Success")

}

func fetchReviewsURL(wg *sync.WaitGroup, lastState *crawler.LastRun) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		new_collection := local_client.Client.Database("docbao").Collection("domain_xe")

		crwl := crawler.InitCrawler()
		for {
			for domain, config := range crwl.WS {
				if config.CategoryType == REVIEW {
					// crawl category - page (pagination)
					links, _ := crwl.FetchSingleURL(domain, config, lastState, PAGE_LIMIT)
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
							created_int := utils.CurrentTimeUnix()
							new := model.News{
								URL:          link,
								CategoryType: config.CategoryType,
								Domain:       domain,
								CreatedInt:   created_int.Unix(),
								Status:       1}
							new.Id = primitive.NewObjectID()
							ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
							_, er := new_collection.InsertOne(ctx, &new)
							if er != nil {
								log.Println("Error, can not insert link", link, er.Error())
								continue
							}
						}
					}
				}

			}

			// break 30 mins before next crawl
			time.Sleep(30 * time.Minute)
			crwl.NewClient()
		}
	}()
}

// crawl a thread/link - include pagination
func crawlURL(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		default_condition := bson.M{
			"category_type": REVIEW,
			"status":        bson.M{"$in": []int{4, 1}}}
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
					var news model.Reviews
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

					crawl_url := news.URL
					u, err := url.Parse(crawl_url)
					if err != nil {
						log.Println("Error, can not get domain, ", err.Error())
					}
					domain := utils.GetDomainName(u.Hostname())
					ws := crwl.WS[domain]
					var cmts = map[string]bool{}

					// initialized last state for fisrst run
					domain_state := crawler.StateDomain{
						CurrentPage: 0,
						ErrCode:     0,
						Status:      true}

					if ws.PaginateRegex != "" {
						for true {

							// request to limit page of thread <= server return 404
							if domain_state.ErrCode == 404 {
								break
							}
							if domain_state.ErrCode < 400 && domain_state.Status {
								domain_state.CurrentPage += 1
							}

							// update crawl_url when page > 1
							pcrawl_url := crawl_url
							if domain_state.CurrentPage > 1 {
								pcrawl_url = fmt.Sprintf(crawl_url+ws.PaginateRegex, domain_state.CurrentPage)
							}

							result, err, res_code := crwl.GetResultCrwl(pcrawl_url)

							// the last comment already exist in map string <= repeat crawl
							// <= server exist to limit page thread but still return 200
							len_reviews := len(result.Reviews)
							if len_reviews == 0 {
								domain_state.Status = false
								domain_state.ErrCode = res_code
								continue
							}

							last_review := result.Reviews[len_reviews-1]
							if _, ok := cmts[last_review]; ok {
								break
							}

							for _, review := range result.Reviews {
								cmts[review] = true
							}

							if err != nil {
								domain_state.Status = false
								domain_state.ErrCode = res_code

								// update proxy status on response_code > 400
								if res_code > 400 && res_code != 404 {
									// update status proxy on init request failed
									proxy_str := strings.Replace(er.Error(), settings.ErrProxyPrefix, "", -1)
									re := regexp.MustCompile(RegexpProxy)

									proxy_pieces := re.FindAllString(strings.TrimSpace(proxy_str), -1)
									if len(proxy_pieces) > 1 {
										ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
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
									}
								}

							} else {
								domain_state.Status = true
								domain_state.ErrCode = 0

								// for first page, only update content
								if domain_state.CurrentPage == 1 {
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
											"date_time":     utils.CurrentTimeUnix().Format("2006-01-02"),
											"updated_str":   utils.CurrentTimeUnix().Format("2006-01-02 15:04:05")}})
									if er != nil {
										log.Println("Error on get content, ", news.Id, er.Error())
									}
								}
							}
						}
					}
					list_review := getListReviews(cmts)
					ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
					_, er = new_collection.UpdateOne(
						ctx,
						bson.M{"_id": news.Id},
						bson.M{"$set": bson.M{
							"reviews":     list_review,
							"status":      2,
							"updated_str": utils.CurrentTimeUnix().Format("2006-01-02 15:04:05")}})
					if er != nil {
						log.Println("Error on get content, ", news.Id, er.Error())
						continue
					}
				}
			}()
		}
		<-done
	}()

}

func getListReviews(cmts map[string]bool) []string {
	result := make([]string, 0)
	for k, _ := range cmts {
		result = append(result, k)
	}

	var final = make([]string, len(result))
	copy(final, result)
	return final
}
