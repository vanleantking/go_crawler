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

	reader "./read_files/structs"

	"gopkg.in/mgo.v2/bson"

	"reflect"

	"./settings"
	"./utils"
)

type Result struct {
	Stocks map[string][]utils.StockInfo
	Err    error
}

var (
	currentDate  = time.Now()
	toDate       = fmt.Sprintf("%d-%d-%d", currentDate.Year(), currentDate.Month(), currentDate.Day())
	fromDate     = "2000-01-01"
	DateRegexp   = `\d{4,}`
	local_client *utils.ClientMGO
	DataTabs     = []string{
		"KQGDThongKeGiaStockPaging",
		"KQGDThongKeDatLenhStockPaging",
		"KQGDGiaoDichNDTNNStockPaging"}
	RefererTabs = []string{
		"thong-ke-gia",
		"thong-ke-lenh",
		"gd-khop-lenh-nn"}

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
	go getPriceDayInfo(&wg)
	// go getOrderMatchInfo(index, crClient, &wg)
	// go getOrderReservationInfo(index, crClient, &wg)
	wg.Wait()
	log.Println("Success")
}

func getPriceDayInfo(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		// initial client custom request
		crClient := settings.NewClient()
		priceday_collection := local_client.Client.Database("new_ck").Collection("price_day")
		header := map[string]string{
			"Accept":                    "*/*",
			"AcceptLanguage":            "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
			"X-Requested-With":          "XMLHttpRequest",
			"Pragma":                    "no-cache",
			"Method":                    "GET",
			"Upgrade-Insecure-Requests": "1"}

		for key := 0; key < len(DataTabs); key++ {
			for _, stockInfos := range Stocks {
				for _, stockInfo := range stockInfos {

					startPage := 1
					maxPage := 1000
					for startPage <= maxPage {
						time.Sleep(utils.RandInRange())
						log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
							DataTabs[key], startPage, utils.PAGESIZE, stockInfo.CatID, stockInfo.StockID, fromDate, toDate)
						fmt.Println("log urllllllllllllllll", log_url, len(stockInfos))

						// setting referer link for each stock
						refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[key], stockInfo.ExchangeCode, stockInfo.StockID)
						header["Referrer"] = refererLink
						fmt.Println("at here", log_url, len(stockInfos), header)

						resp, err := crClient.InitCustomRequest(log_url, header)
						fmt.Println("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", resp, err)
						if err != nil {
							fmt.Println(err.Error())
						}
						defer resp.Body.Close()
						body, err := ioutil.ReadAll(resp.Body)
						fmt.Println("errrorrrrrrrrrrrr", string(body), err)
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
												fmt.Println("price day, ", priceDay)
												ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
												_, er = priceday_collection.InsertOne(ctx, &priceDay)
												if er != nil {
													log.Println("Error on insert last_day, ", priceDayInterface)
												}
											} else {
												break
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
		}
		defer wg.Done()
	}()
}

func getOrderMatchInfo(key int, crClient *settings.Client, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		matchorder_collection := local_client.Client.Database("new_ck").Collection("match_order")
		header := map[string]string{
			"Accept":                    "*/*",
			"AcceptLanguage":            "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
			"X-Requested-With":          "XMLHttpRequest",
			"Pragma":                    "no-cache",
			"Method":                    "GET",
			"Upgrade-Insecure-Requests": "1"}

		for _, stockInfos := range Stocks {
			for _, stockInfo := range stockInfos {

				startPage := 1
				maxPage := 1000
				for startPage <= maxPage {
					time.Sleep(utils.RandInRange())
					log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
						DataTabs[key], startPage, utils.PAGESIZE, stockInfo.CatID, stockInfo.StockID, fromDate, toDate)
					fmt.Println(log_url)

					// setting referer link for each stock
					refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[key], stockInfo.ExchangeCode, stockInfo.StockCode)
					header["Referrer"] = refererLink

					resp, err := crClient.InitCustomRequest(log_url, header)
					if err != nil {
						fmt.Println(err.Error())
					}
					defer resp.Body.Close()
					body, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						fmt.Println(err.Error())
					}
					fmt.Println("boy dy string", string(body))
					var result []interface{}
					err = json.Unmarshal(body, &result)
					if err != nil {
						fmt.Println(err.Error())
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
									fmt.Println("order match interface, ", orderMatchInterface)
									matchDate := regexp.MustCompile(DateRegexp)
									trDate := int64(0)
									if matchDate.MatchString(orderMatchInterface["TradingDate"].(string)) {
										match, _ := strconv.ParseInt(matchDate.FindString(orderMatchInterface["TradingDate"].(string)), 10, 64)
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
											matchOrder.TradingDate = trDate // time in nano-second
											matchOrder.StockCode = stockInfo.StockCode
											matchOrder.StockID = stockInfo.StockID
											matchOrder.TrID = orderMatchInterface["TrID"].(float64)
											matchOrder.TotalVol = orderMatchInterface["TotalVol"].(float64)
											matchOrder.TotalVal = orderMatchInterface["TotalVal"].(float64)
											matchOrder.TotalRoom = orderMatchInterface["TotalRoom"].(float64)
											matchOrder.CurrRoom = orderMatchInterface["CurrRoom"].(float64)
											matchOrder.BuyVol = orderMatchInterface["BuyVol"].(float64)
											matchOrder.PerBuyVol = orderMatchInterface["PerBuyVol"].(float64)
											matchOrder.BuyVal = orderMatchInterface["BuyVal"].(float64)
											matchOrder.PerBuyVal = orderMatchInterface["PerBuyVal"].(float64)
											matchOrder.SellVol = orderMatchInterface["SellVol"].(float64)
											matchOrder.PerSellVol = orderMatchInterface["PerSellVol"].(float64)
											matchOrder.SellVal = orderMatchInterface["SellVal"].(float64)
											matchOrder.PerSellVal = orderMatchInterface["PerSellVal"].(float64)
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
											fmt.Println("match order, ", matchOrder)
											ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
											_, er = matchorder_collection.InsertOne(ctx, &matchOrder)
											if er != nil {
												log.Println("Error on insert order_match, ", orderMatchInterface)
											}
										} else {
											break
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

func getOrderReservationInfo(key int, crClient *settings.Client, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		reservationorder_collection := local_client.Client.Database("new_ck").Collection("reservation_order")
		header := map[string]string{
			"Accept":                    "*/*",
			"AcceptLanguage":            "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
			"X-Requested-With":          "XMLHttpRequest",
			"Pragma":                    "no-cache",
			"Method":                    "GET",
			"Upgrade-Insecure-Requests": "1"}

		for _, stockInfos := range Stocks {
			for _, stockInfo := range stockInfos {

				startPage := 1
				maxPage := 1000
				for startPage <= maxPage {
					fmt.Println("start page, ", startPage, maxPage, len(stockInfos))
					time.Sleep(utils.RandInRange())
					log_url := fmt.Sprintf(utils.VIETSTOCK_DATA,
						DataTabs[key], startPage, utils.PAGESIZE, stockInfo.CatID, stockInfo.StockID, fromDate, toDate)
					fmt.Println("log url, ", log_url)

					// setting referer link for each stock
					refererLink := fmt.Sprintf(utils.STOCK_REFERER, RefererTabs[key], stockInfo.ExchangeCode, stockInfo.StockCode)
					header["Referrer"] = refererLink

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
					fmt.Println("resultttttttttttttttttttt", result)

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
											reserverOrder.TradingDate = trDate // time in nano-second
											reserverOrder.StockCode = stockInfo.StockCode
											reserverOrder.StockID = stockInfo.StockID
											reserverOrder.OutstandingBuy = orderReserveInterface["OutstandingBuy"].(float64)
											reserverOrder.OutstandingSell = orderReserveInterface["OutstandingSell"].(float64)
											reserverOrder.TotalBuyTrade = orderReserveInterface["TotalBuyTrade"].(float64)
											reserverOrder.TotalBuyVol = orderReserveInterface["TotalBuyVol"].(float64)
											reserverOrder.TotalSellTrade = orderReserveInterface["TotalSellTrade"].(float64)
											reserverOrder.TotalSellVol = orderReserveInterface["TotalSellVol"].(float64)
											reserverOrder.TrID = orderReserveInterface["TrID"].(float64)
											reserverOrder.DisparityTrade = orderReserveInterface["DisparityTrade"].(float64)
											reserverOrder.DisparityVol = orderReserveInterface["DisparityVol"].(float64)
											reserverOrder.AVGVolBuy = orderReserveInterface["AVGVolBuy"].(float64)
											reserverOrder.AVGVolSell = orderReserveInterface["AVGVolSell"].(float64)
											reserverOrder.ExchangeCode = stockInfo.ExchangeCode
											reserverOrder.ExchangeName = stockInfo.ExchangeName
											trStr := time.Unix(int64(trDate/1000), 0)
											reserverOrder.TradingDateStr = strings.Split(trStr.Add(7*time.Hour).Format("2006-01-02 15:04:05"), " ")[0]
											fmt.Println("price day, ", reserverOrder)
											ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
											_, er = reservationorder_collection.InsertOne(ctx, &reserverOrder)
											if er != nil {
												log.Println("Error on insert order_match, ", orderReserveInterface)
											}
										} else {
											break
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
