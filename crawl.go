package main

// implement client request to get data from vietstock
import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/mgo.v2/bson"

	"reflect"

	"./settings"
	"./utils"
)

var (
	currentDate  = time.Now()
	toDate       = fmt.Sprintf("%d-%d-%d", currentDate.Year(), currentDate.Month(), currentDate.Day())
	fromDate     = "2019-01-01"
	DateRegexp   = `\d{4,}`
	local_client *utils.ClientMGO

	// Stock = map[string][]StockInfo{
	// 	"HOSE": []StockInfo{
	// 		StockInfo{CatID: 1, StockID: 497, MaxPage: 0, StockCode: "AAA"}}}
	// ,
	// StockInfo{CatID: 1, StockID: -19, MaxPage: 0, StockCode: "VN-Index"}}}
	Stock = map[string][]StockInfo{
		"VN30": []StockInfo{
			StockInfo{CatID: 4, StockID: -16, MaxPage: 0, StockCode: "VN30-Index"}}}
)

//StockInfo{CatStock: 1, StockID: 497, StockCode: "AAA", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"},
//StockInfo{CatStock: 1, StockID: -19, StockCode: "VN-Index", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"}

type StockInfo struct {
	CatID     int
	StockID   int
	MaxPage   int
	StockCode string
}

type LastDay struct {
	CloseIndex  float64 `json:"CloseIndex" bson:"CloseIndex"`
	PriorIndex  float64 `json:"PriorIndex" bson:"PriorIndex"`
	Change      float64 `json:"Change" bson:"Change"`
	PerChange   float64 `json:"PerChange" bson:"PerChange"`
	ChangeColor string  `json:"ChangeColor" bson:"ChangeColor"`
	ChangeText  string  `json:"ChangeText" bson:"ChangeText"`
	TrDate      int64   `json:"TrDate" bson:"TrDate"`
	TranNo      float64 `json:"TranNo" bson:"TranNo"`
	StockCode   string  `json:"StockCode" bson:"StockCode"`
	TrDateStr   string  `json:"TrDateStr" bson:"TrDateStr"`
}

type PriceDay struct {
	TradingDate  int64   `json:"TradingDate" bson:"TradingDate"`
	StockCode    string  `json:"StockCode" bson:"StockCode"`
	FinanceURL   string  `json:"FinanceURL" bson:"FinanceURL"`
	StockName    string  `json:"StockName" bson:"StockName"`
	BasicPrice   float64 `json:"BasicPrice" bson:"BasicPrice"`
	OpenPrice    float64 `json:"OpenPrice" bson:"OpenPrice"`
	ClosePrice   float64 `json:"ClosePrice" bson:"ClosePrice"`
	HighestPrice float64 `json:"HighestPrice" bson:"HighestPrice"`
	LowestPrice  float64 `json:"LowestPrice" bson:"LowestPrice"`
	AvrPrice     float64 `json:"AvrPrice" bson:"AvrPrice"`
	Change       float64 `json:"Change" bson:"Change"`
	PerChange    float64 `json:"PerChange" bson:"PerChange"`
	ChangeColor  string  `json:"ChangeColor" bson:"ChangeColor"`
	ChangeImage  string  `json:"ChangeImage" bson:"ChangeImage"`
	M_TotalVol   float64 `json:"M_TotalVol" bson:"M_TotalVol"`
	M_TotalVal   float64 `json:"M_TotalVal" bson:"M_TotalVal"`
	PT_TotalVol  float64 `json:"PT_TotalVol" bson:"PT_TotalVol"`
	PT_TotalVal  float64 `json:"PT_TotalVal" bson:"PT_TotalVal"`
	TotalVol     float64 `json:"TotalVol" bson:"TotalVol"`
	TotalVal     float64 `json:"TotalVal" bson:"TotalVal"`
	MarketCap    float64 `json:"MarketCap" bson:"MarketCap"`
	StockNameEn  string  `json:"StockNameEn" bson:"StockNameEn"`
	ROW          float64 `json:"ROW" bson:"ROW"`
	StockID      float64 `json:"StockID" bson:"StockID"`
	TrID         float64 `json:"TrID" bson:"TrID"`
	TrDateStr    string  `json:"TrDateStr" bson:"TrDateStr"`
}

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/ck.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
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

	lastday_collection := local_client.Client.Database("ck").Collection("last_day")
	priceday_collection := local_client.Client.Database("ck").Collection("price_day")

	// initial client custom request
	crClient := settings.NewClient()

	header := map[string]string{
		"Referrer":                  "https://finance.vietstock.vn/ket-qua-giao-dich?tab=thong-ke-gia&exchange=1&code=-16",
		"Accept":                    "*/*",
		"AcceptLanguage":            "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		"X-Requested-With":          "XMLHttpRequest",
		"Pragma":                    "no-cache",
		"Method":                    "GET",
		"Upgrade-Insecure-Requests": "1"}

	//STOCK_PARAMETER {page, pageSize, catID, stockID, fromDate, toDate}
	for _, stockInfos := range Stock {
		for _, stockInfo := range stockInfos {

			startPage := 1
			maxPage := 1000
			for startPage <= maxPage {
				parameters := fmt.Sprintf(utils.STOCK_PARAMETER,
					startPage, utils.PAGESIZE, stockInfo.CatID, stockInfo.StockID, fromDate, toDate)
				log_url := strings.TrimSpace(utils.VIETSTOCK_BASEURL + parameters)
				fmt.Println(log_url)
				resp, err := crClient.InitCustomRequest(log_url, header)
				if err != nil {
					fmt.Println(err.Error())
				}
				defer resp.Body.Close()
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Println(err.Error())
				}
				var result []interface{}
				err = json.Unmarshal(body, &result)
				if err != nil {
					fmt.Println(err.Error())
				}

				for key, re := range result {
					switch key {
					// last day type
					case 0:
						switch reflect.TypeOf(re).Kind() {
						case reflect.Slice:
							tmp_slice := reflect.ValueOf(re)
							for i := 0; i < tmp_slice.Len(); i++ {
								lastDayInterface := tmp_slice.Index(i).Interface().(map[string]interface{})
								// regex date from json return
								matchDate := regexp.MustCompile(DateRegexp)
								trDate := int64(0)
								if matchDate.MatchString(lastDayInterface["TrDate"].(string)) {
									match, _ := strconv.ParseInt(matchDate.FindString(lastDayInterface["TrDate"].(string)), 10, 64)
									trDate = match
								}
								ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
								count, er := lastday_collection.Count(
									ctx,
									bson.M{
										"StockCode": stockInfo.StockCode,
										"TrDate":    trDate})
								if er != nil {
									log.Println("Error on count last_day, ", lastDayInterface)
								} else {
									// only insert if count == 0
									if count == int64(0) {
										lastDay := LastDay{}
										lastDay.TrDate = trDate // time in nano-second
										lastDay.Change = lastDayInterface["Change"].(float64)
										lastDay.PerChange = lastDayInterface["PerChange"].(float64)
										lastDay.ChangeColor = lastDayInterface["ChangeColor"].(string)
										lastDay.ChangeText = lastDayInterface["ChangeText"].(string)
										lastDay.TranNo = lastDayInterface["TranNo"].(float64)
										lastDay.CloseIndex = lastDayInterface["CloseIndex"].(float64)
										lastDay.PriorIndex = lastDayInterface["PriorIndex"].(float64)
										lastDay.StockCode = stockInfo.StockCode
										trStr := time.Unix(int64(trDate/1000), 0)
										lastDay.TrDateStr = trStr.Add(7 * time.Hour).Format("2006-01-02 15:04:05")
										ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
										_, er = lastday_collection.InsertOne(ctx, &lastDay)
										if er != nil {
											log.Println("Error on insert last_day, ", lastDayInterface)
										}
									}
								}
							}
						}
					// Price day
					case 1:
						switch reflect.TypeOf(re).Kind() {
						case reflect.Slice:
							tmp_slice := reflect.ValueOf(re)
							for i := 0; i < tmp_slice.Len(); i++ {
								priceDayInterface := tmp_slice.Index(i).Interface().(map[string]interface{})
								matchDate := regexp.MustCompile(DateRegexp)
								trDate := int64(0)
								if matchDate.MatchString(priceDayInterface["TradingDate"].(string)) {
									match, _ := strconv.ParseInt(matchDate.FindString(priceDayInterface["TradingDate"].(string)), 10, 64)
									trDate = match
								}

								ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
								count, er := priceday_collection.Count(
									ctx,
									bson.M{
										"TradingDate": trDate,
										"StockCode":   stockInfo.StockCode})
								if er != nil {
									log.Println("Error on count last_day, ", priceDayInterface)
								} else {
									// only insert if count == 0
									if count == int64(0) {
										priceDay := PriceDay{}
										priceDay.TradingDate = trDate // time in nano-second
										priceDay.StockCode = priceDayInterface["StockCode"].(string)
										priceDay.FinanceURL = priceDayInterface["FinanceURL"].(string)
										priceDay.StockName = priceDayInterface["StockName"].(string)
										priceDay.BasicPrice = priceDayInterface["BasicPrice"].(float64)
										priceDay.OpenPrice = priceDayInterface["OpenPrice"].(float64)
										priceDay.ClosePrice = priceDayInterface["ClosePrice"].(float64)
										priceDay.HighestPrice = priceDayInterface["HighestPrice"].(float64)
										priceDay.LowestPrice = priceDayInterface["LowestPrice"].(float64)
										priceDay.AvrPrice = priceDayInterface["AvrPrice"].(float64)
										priceDay.Change = priceDayInterface["Change"].(float64)
										priceDay.PerChange = priceDayInterface["PerChange"].(float64)
										priceDay.ChangeColor = priceDayInterface["ChangeColor"].(string)
										priceDay.ChangeImage = priceDayInterface["ChangeImage"].(string)
										priceDay.M_TotalVol = priceDayInterface["M_TotalVol"].(float64)
										priceDay.M_TotalVal = priceDayInterface["M_TotalVal"].(float64)
										priceDay.PT_TotalVol = priceDayInterface["PT_TotalVol"].(float64)
										priceDay.PT_TotalVal = priceDayInterface["PT_TotalVal"].(float64)
										priceDay.TotalVol = priceDayInterface["TotalVol"].(float64)
										priceDay.TotalVal = priceDayInterface["TotalVal"].(float64)
										priceDay.MarketCap = priceDayInterface["MarketCap"].(float64)
										priceDay.StockNameEn = priceDayInterface["StockNameEn"].(string)
										priceDay.ROW = priceDayInterface["ROW"].(float64)
										priceDay.StockID = priceDayInterface["StockID"].(float64)
										priceDay.TrID = priceDayInterface["TrID"].(float64)
										trStr := time.Unix(int64(trDate/1000), 0)
										priceDay.TrDateStr = trStr.Add(7 * time.Hour).Format("2006-01-02 15:04:05")
										fmt.Println("price day, ", priceDay)
										ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
										_, er = priceday_collection.InsertOne(ctx, &priceDay)
										if er != nil {
											log.Println("Error on insert last_day, ", priceDayInterface)
										}
									}
								}
							}
						}
					// get max page pagination
					case 2:
						switch reflect.TypeOf(re).Kind() {
						case reflect.Slice:
							tmp_slice := reflect.ValueOf(re)
							for i := 0; i < tmp_slice.Len(); i++ {
								priceDayInterface := tmp_slice.Index(i).Interface().(float64)
								maxPage = int(priceDayInterface)
								break
							}

						}
					}
				}
				startPage++
			}
		}
	}
	log.Println("Success")
}
