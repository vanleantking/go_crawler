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
	"sync"
	"time"

	"github.com/mongodb/mongo-go-driver/bson/primitive"

	reader "./read_files/structs"

	"gopkg.in/mgo.v2/bson"

	"reflect"

	"../settings"
	"../utils"
)

type Result struct {
	Stocks map[string][]utils.StockInfo
	Err    error
}

var (
	currentDate = time.Now()

	fromDate     = "2019-05-26"
	DateRegexp   = `\d{4,}`
	local_client *utils.ClientMGO
	DataTabs     = []string{
		"KQGDThongKeGiaStockPaging",
		"KQGDGiaoDichNDTNNStockPaging",
		"KQGDThongKeDatLenhStockPaging"}
	RefererTabs = []string{
		"thong-ke-gia",
		"thong-ke-lenh",
		"gd-khop-lenh-nn"}
	header = map[string]string{
		"Cache-Control":             "max-age=0",
		"Accept":                    "*/*",
		"AcceptLanguage":            "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		"Method":                    "GET",
		"Upgrade-Insecure-Requests": "1"}

	result = readData()
	Stocks map[string][]utils.StockInfo
)

//StockInfo{CatStock: 1, StockID: 497, StockCode: "AAA", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"},
//StockInfo{CatStock: 1, StockID: -19, StockCode: "VN-Index", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"}

func main() {
	var wg sync.WaitGroup
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
	// wg.Add(1)

	if result.Err != nil {
		fmt.Println("Opps. something happened in read file", err.Error())
		return
	}
	Stocks = result.Stocks

	var er error
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["localhost"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)

	// lastday_collection := local_client.Client.Database("new_ck").Collection("last_day")

	//STOCK_PARAMETER {page, pageSize, catID, stockID, fromDate, toDate}
	for true {
		toDate := fmt.Sprintf("%d-%d-%d", currentDate.Year(), currentDate.Month(), currentDate.Day())
		// fromDate := toDate
		getPriceDayInfo(fromDate, toDate, &wg)
		getOrderMatchInfo(fromDate, toDate, &wg)
		getOrderReservationInfo(fromDate, toDate, &wg)
		// break 24 hour for next crawl
		log.Println("Info, Break 24 hours")
		time.Sleep(24 * time.Hour)
	}
	wg.Wait()
	log.Println("Success")
}

func getPriceDayInfo(fromDate string, toDate string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// initial client custom request
		crClient := settings.NewClient()
		priceday_collection := local_client.Client.Database("new_ck").Collection("price_day")

		for _, stockInfos := range Stocks {
			for _, stockInfo := range stockInfos {
				startPage := 1
				maxPage := 1000
				for startPage <= maxPage {
					time.Sleep(utils.RangeWideTimeOut())
					log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
						DataTabs[0], startPage, utils.PAGESIZE, stockInfo.CatID,
						stockInfo.StockID, fromDate, toDate)

					// setting referer link for each stock
					refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[0],
						stockInfo.ExchangeCode, stockInfo.StockID)
					header["Referrer"] = refererLink

					resp, err := crClient.InitCustomRequest(log_url, header)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					var result []interface{}
					err = json.Unmarshal(body, &result)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}

					for key, re := range result {
						switch key {
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
										// only insert if count == 0 else break <= save already
										if count == int64(0) {
											priceDay := utils.PriceDay{}
											priceDay.Id = primitive.NewObjectID()
											priceDay.TradingDate = trDate // time in nano-second
											priceDay.StockCode = priceDayInterface["StockCode"].(string)
											priceDay.BasicPrice = priceDayInterface["BasicPrice"].(float64)
											priceDay.OpenPrice = priceDayInterface["OpenPrice"].(float64)
											priceDay.ClosePrice = priceDayInterface["ClosePrice"].(float64)
											priceDay.HighestPrice = priceDayInterface["HighestPrice"].(float64)
											priceDay.LowestPrice = priceDayInterface["LowestPrice"].(float64)
											priceDay.AvrPrice = priceDayInterface["AvrPrice"].(float64)
											priceDay.Change = priceDayInterface["Change"].(float64)
											priceDay.PerChange = priceDayInterface["PerChange"].(float64)
											priceDay.ChangeText = priceDayInterface["ChangeColor"].(string)
											priceDay.M_TotalVol = priceDayInterface["M_TotalVol"].(float64)
											priceDay.M_TotalVal = priceDayInterface["M_TotalVal"].(float64)
											priceDay.PT_TotalVol = priceDayInterface["PT_TotalVol"].(float64)
											priceDay.PT_TotalVal = priceDayInterface["PT_TotalVal"].(float64)
											priceDay.TotalVol = priceDayInterface["TotalVol"].(float64)
											priceDay.TotalVal = priceDayInterface["TotalVal"].(float64)
											priceDay.MarketCap = priceDayInterface["MarketCap"].(float64)
											priceDay.StockID = stockInfo.StockID
											priceDay.TrID = priceDayInterface["TrID"].(float64)
											priceDay.ExchangeCode = stockInfo.ExchangeCode
											priceDay.ExchangeName = stockInfo.ExchangeName
											trStr := time.Unix(int64(trDate/1000), 0)
											priceDay.TrDateStr = strings.Split(trStr.Add(7*time.Hour).Format("2006-01-02 15:04:05"), " ")[0]
											ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
											_, er = priceday_collection.InsertOne(ctx, &priceDay)
											if er != nil {
												log.Println("Error on insert last_day, ", er, priceDayInterface)
											}
										} else {
											continue
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
									continue
								}
							}
						}
					}
					startPage++
				}
			}
		}
	}()
}

func getOrderMatchInfo(fromDate string, toDate string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		matchorder_collection := local_client.Client.Database("new_ck").Collection("match_order")

		for _, stockInfos := range Stocks {
			for _, stockInfo := range stockInfos {
				// initial client custom request
				crClient := settings.NewClient()
				startPage := 1
				maxPage := 1000
				for startPage <= maxPage {
					time.Sleep(utils.RangeWideTimeOut())
					log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
						DataTabs[1], startPage, utils.PAGESIZE, stockInfo.CatID,
						stockInfo.StockID, fromDate, toDate)

					// setting referer link for each stock
					refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[1],
						stockInfo.ExchangeCode, stockInfo.StockCode)
					header["Referrer"] = refererLink

					resp, err := crClient.InitCustomRequest(log_url, header)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					var result []interface{}
					err = json.Unmarshal(body, &result)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}

					for key, re := range result {
						switch key {
						// Price day
						case 1:
							switch reflect.TypeOf(re).Kind() {
							case reflect.Slice:
								tmp_slice := reflect.ValueOf(re)
								for i := 0; i < tmp_slice.Len(); i++ {
									orderMatchInterface := tmp_slice.Index(i).Interface().(map[string]interface{})
									matchDate := regexp.MustCompile(DateRegexp)
									trDate := int64(0)
									if matchDate.MatchString(orderMatchInterface["TradingDate"].(string)) {
										match, _ := strconv.ParseInt(matchDate.FindString(
											orderMatchInterface["TradingDate"].(string)), 10, 64)
										trDate = match
									}

									ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
									count, er := matchorder_collection.Count(
										ctx,
										bson.M{
											"TradingDate": trDate,
											"StockCode":   stockInfo.StockCode})
									if er != nil {
										log.Println("Error on count oder_match, ", orderMatchInterface)
									} else {
										// only insert if count == 0 else break <= save already
										if count == int64(0) {
											matchOrder := utils.MatchOrder{}
											matchOrder.Id = primitive.NewObjectID()
											matchOrder.TradingDate = trDate // time in nano-second
											matchOrder.StockCode = stockInfo.StockCode
											matchOrder.StockID = stockInfo.StockID
											matchOrder.TotalRoom = orderMatchInterface["TotalRoom"].(float64)
											matchOrder.CurrRoom = orderMatchInterface["CurrRoom"].(float64)
											matchOrder.RemainRoom = orderMatchInterface["RemainRoom"].(float64)
											matchOrder.OwnedRatio = orderMatchInterface["OwnedRatio"].(float64)
											matchOrder.DiffBuySellPutVol = orderMatchInterface["DiffBuySellPutVol"].(float64)
											matchOrder.BuyVol = orderMatchInterface["BuyVol"].(float64)
											matchOrder.PerBuyVol = orderMatchInterface["PerBuyVol"].(float64)
											matchOrder.BuyVal = orderMatchInterface["BuyVal"].(float64)
											matchOrder.PerBuyVal = orderMatchInterface["PerBuyVal"].(float64)
											matchOrder.SellVol = orderMatchInterface["SellVol"].(float64)
											matchOrder.PerSellVol = orderMatchInterface["PerSellVol"].(float64)
											matchOrder.SellVal = orderMatchInterface["SellVal"].(float64)
											matchOrder.PerSellVal = orderMatchInterface["PerSellVal"].(float64)
											matchOrder.DiffBuySellPutVal = orderMatchInterface["DiffBuySellPutVal"].(float64)
											matchOrder.DiffBuySellVol = orderMatchInterface["DiffBuySellVol"].(float64)
											matchOrder.DiffBuySellVal = orderMatchInterface["DiffBuySellVal"].(float64)
											matchOrder.BuyPutVol = orderMatchInterface["BuyPutVol"].(float64)
											matchOrder.PerBuyPutVol = orderMatchInterface["PerBuyPutVol"].(float64)
											matchOrder.BuyPutVal = orderMatchInterface["BuyPutVal"].(float64)
											matchOrder.PerBuyPutVal = orderMatchInterface["PerBuyPutVal"].(float64)
											matchOrder.SellPutVol = orderMatchInterface["SellPutVol"].(float64)
											matchOrder.PerSellPutVol = orderMatchInterface["PerSellPutVol"].(float64)
											matchOrder.SellPutVal = orderMatchInterface["SellPutVal"].(float64)
											matchOrder.PerSellPutVal = orderMatchInterface["PerSellPutVal"].(float64)
											matchOrder.ExchangeCode = stockInfo.ExchangeCode
											matchOrder.ExchangeName = stockInfo.ExchangeName
											trStr := time.Unix(int64(trDate/1000), 0)
											matchOrder.TradingDateStr = strings.Split(trStr.Add(7*time.Hour).Format("2006-01-02 15:04:05"), " ")[0]

											ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
											_, er = matchorder_collection.InsertOne(ctx, &matchOrder)
											if er != nil {
												log.Println("Error on insert order_match, ", er, orderMatchInterface)
											}
										} else {
											continue
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
									continue
								}
							}
						}
					}
					startPage++
				}
			}
		}
	}()
}

func getOrderReservationInfo(fromDate string, toDate string, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		reservationorder_collection := local_client.Client.Database("new_ck").Collection("reservation_order")
		for _, stockInfos := range Stocks {
			for _, stockInfo := range stockInfos {
				// initial client custom request
				crClient := settings.NewClient()
				startPage := 1
				maxPage := 1000
				for startPage <= maxPage {
					time.Sleep(utils.RangeWideTimeOut())
					log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
						DataTabs[2], startPage, utils.PAGESIZE, stockInfo.CatID, stockInfo.StockID, fromDate, toDate)

					// setting referer link for each stock
					refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[2], stockInfo.ExchangeCode, stockInfo.StockCode)
					header["Referrer"] = refererLink

					resp, err := crClient.InitCustomRequest(log_url, header)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}
					var result []interface{}
					err = json.Unmarshal(body, &result)
					if err != nil {
						log.Println(err.Error(), stockInfo, fromDate, toDate)
						continue
					}

					for key, re := range result {
						switch key {
						// Price day
						case 1:
							switch reflect.TypeOf(re).Kind() {
							case reflect.Slice:
								tmp_slice := reflect.ValueOf(re)
								for i := 0; i < tmp_slice.Len(); i++ {
									orderReserveInterface := tmp_slice.Index(i).Interface().(map[string]interface{})
									matchDate := regexp.MustCompile(DateRegexp)
									trDate := int64(0)
									if matchDate.MatchString(orderReserveInterface["TradingDate"].(string)) {
										match, _ := strconv.ParseInt(matchDate.FindString(orderReserveInterface["TradingDate"].(string)), 10, 64)
										trDate = match
									}

									ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
									count, er := reservationorder_collection.Count(
										ctx,
										bson.M{
											"TradingDate": trDate,
											"StockCode":   stockInfo.StockCode})
									if er != nil {
										log.Println("Error on count oder_match, ", orderReserveInterface)
									} else {
										// only insert if count == 0 else break <= save already
										if count == int64(0) {
											reserverOrder := utils.ReserveOrder{}
											reserverOrder.Id = primitive.NewObjectID()
											reserverOrder.TradingDate = trDate // time in nano-second
											reserverOrder.StockCode = stockInfo.StockCode
											reserverOrder.StockID = stockInfo.StockID
											reserverOrder.ExchangeCode = stockInfo.ExchangeCode
											reserverOrder.ExchangeName = stockInfo.ExchangeName
											reserverOrder.ClosePrice = orderReserveInterface["ClosePrice"].(float64)
											reserverOrder.TotalVol = orderReserveInterface["TotalVol"].(float64)
											reserverOrder.TotalVal = orderReserveInterface["TotalVal"].(float64)
											reserverOrder.BestBuy = orderReserveInterface["BestBuy"].(float64)
											reserverOrder.BestBidVol = orderReserveInterface["BestBidVol"].(float64)
											reserverOrder.BestSell = orderReserveInterface["BestSell"].(float64)
											reserverOrder.BestSellVol = orderReserveInterface["BestSellVol"].(float64)
											reserverOrder.TotalBuyTrade = orderReserveInterface["TotalBuyTrade"].(float64)
											reserverOrder.TotalSellTrade = orderReserveInterface["TotalSellTrade"].(float64)
											reserverOrder.SpreadBSTrade = orderReserveInterface["SpreadBSTrade"].(float64)
											reserverOrder.TotalBuyVol = orderReserveInterface["TotalBuyVol"].(float64)
											reserverOrder.TotalSellVol = orderReserveInterface["TotalSellVol"].(float64)
											reserverOrder.SpreadBSVol = orderReserveInterface["SpreadBSVol"].(float64)
											trStr := time.Unix(int64(trDate/1000), 0)
											reserverOrder.TradingDateStr = strings.Split(trStr.Add(7*time.Hour).Format("2006-01-02 15:04:05"), " ")[0]
											ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
											_, er = reservationorder_collection.InsertOne(ctx, &reserverOrder)
											if er != nil {
												log.Println("Error on insert order_match, ", er, orderReserveInterface)
											}
										} else {
											continue
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
	}()
}

func readData() Result {

	var stockinfos = map[string][]utils.StockInfo{}
	extension := "xlsx"

	reader := &reader.Reader{Path: "./read_files/data/data.xlsx"}
	reader.SetType(extension)
	err := reader.GetData()
	if err != nil {
		return Result{Stocks: stockinfos, Err: err}
	}
	reader.GetHeaderFileImport()
	stocks := make([]utils.StockInfo, 0, len(reader.Rows)-1)
	for i := 1; i < len(reader.Rows); i++ {
		stockInfo := utils.StockInfo{}
		getStockByReader(&stockInfo, reader.Rows[i], reader)
		stocks = append(stocks, stockInfo)
	}
	stockinfos["VN30"] = stocks
	return Result{Stocks: stockinfos, Err: nil}
}

func getStockByReader(stockInfo *utils.StockInfo, data []string, reader *reader.Reader) {
	for key, index := range reader.Header {
		switch key {
		case "CatID":
			cat, _ := strconv.Atoi(data[index])
			stockInfo.CatID = cat
		case "StockID":
			cat, _ := strconv.Atoi(data[index])
			stockInfo.StockID = cat
		case "ExchangeCode":
			cat, _ := strconv.Atoi(data[index])
			stockInfo.ExchangeCode = cat
		case "StockCode":
			stockInfo.StockCode = data[index]
		case "ExchangeName":
			stockInfo.ExchangeName = data[index]
		}
	}
}
