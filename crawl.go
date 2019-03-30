package main

// implement client request to get data from vietstock
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"reflect"

	"./settings"
	"./utils"
)

var (
	currentDate      = time.Now()
	toDate           = fmt.Sprintf("%d-%d-%d", currentDate.Year(), currentDate.Month(), currentDate.Day())
	IDStatisticPrice = "#statistic-price .table tbody"
	IDTabView        = "view-tab"
	Result           = map[string][]string{}
	fromDate         = "2001-01-01"
	DateRegexp       = `\d{4,}`

	Stock = map[string][]StockInfo{
		"HOSE": []StockInfo{
			StockInfo{CatID: 1, StockID: 497, MaxPage: 0},
			StockInfo{CatID: 1, StockID: -19, MaxPage: 0}}}
)

//StockInfo{CatStock: 1, StockID: 497, StockCode: "AAA", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"},
//StockInfo{CatStock: 1, StockID: -19, StockCode: "VN-Index", StockName: "CTCP Nhựa và Môi trường Xanh An Phát"}

type StockInfo struct {
	CatID   int
	StockID int
	MaxPage int
}

type LastDay struct {
	CloseIndex  float64 `json:"CloseIndex"`
	PriorIndex  float64 `json:"PriorIndex"`
	Change      float64 `json:"Change"`
	PerChange   float64 `json:"PerChange"`
	ChangeColor string  `json:"ChangeColor"`
	ChangeText  string  `json:"ChangeText"`
	TrDate      int64   `json:"TrDate"`
	TranNo      float64 `json:"TranNo"`
}

type PriceDay struct {
	TradingDate  int64   `json:"TradingDate"`
	StockCode    string  `json:"StockCode"`
	FinanceURL   string  `json:"FinanceURL"`
	StockName    string  `json:"StockName"`
	BasicPrice   int64   `json:"BasicPrice"`
	OpenPrice    int64   `json:"OpenPrice"`
	ClosePrice   int64   `json:"ClosePrice"`
	HighestPrice int64   `json:"HighestPrice"`
	LowestPrice  int64   `json:"LowestPrice"`
	AvrPrice     int64   `json:"AvrPrice"`
	Change       int64   `json:"Change"`
	PerChange    float64 `json:"PerChange"`
	ChangeColor  string  `json:"ChangeColor"`
	ChangeImage  string  `json:"ChangeImage"`
	M_TotalVol   int64   `json:"M_TotalVol"`
	M_TotalVal   float64 `json:"M_TotalVal"`
	PT_TotalVol  float64 `json:"PT_TotalVol"`
	PT_TotalVal  float64 `json:"PT_TotalVal"`
	TotalVol     int64   `json:"TotalVol"`
	TotalVal     float64 `json:"TotalVal"`
	MarketCap    float64 `json:"MarketCap"`
	StockNameEn  string  `json:"StockNameEn"`
	ROW          int     `json:"ROW"`
	StockID      int     `json:"StockID"`
	TrID         int     `json:"TrID"`
}

func main() {
	// initial client custom request
	crClient := settings.NewClient()

	header := map[string]string{
		"Referrer":                  "https://finance.vietstock.vn/ket-qua-giao-dich?tab=thong-ke-gia&exchange=1&code=-19",
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
							lastDay := LastDay{}
							lastDay.Change = lastDayInterface["Change"].(float64)
							lastDay.PerChange = lastDayInterface["PerChange"].(float64)
							lastDay.ChangeColor = lastDayInterface["ChangeColor"].(string)
							lastDay.ChangeText = lastDayInterface["ChangeText"].(string)
							matchDate := regexp.MustCompile(DateRegexp)
							if matchDate.MatchString(lastDayInterface["TrDate"].(string)) {
								match, _ := strconv.ParseInt(matchDate.FindString(lastDayInterface["TrDate"].(string)), 10, 64)
								lastDay.TrDate = match
							}
							lastDay.TranNo = lastDayInterface["TranNo"].(float64)
							lastDay.CloseIndex = lastDayInterface["CloseIndex"].(float64)
							lastDay.PriorIndex = lastDayInterface["PriorIndex"].(float64)
							fmt.Println(lastDay)
						}
					}
				// Price day
				case 1:
					switch reflect.TypeOf(re).Kind() {
					case reflect.Slice:
						tmp_slice := reflect.ValueOf(re)
						for i := 0; i < tmp_slice.Len(); i++ {
							priceDayInterface := tmp_slice.Index(i).Interface().(map[string]interface{})
							priceDay := PriceDay{}
							lastDay.Change = lastDayInterface["Change"].(float64)
							lastDay.PerChange = lastDayInterface["PerChange"].(float64)
							lastDay.ChangeColor = lastDayInterface["ChangeColor"].(string)
							lastDay.ChangeText = lastDayInterface["ChangeText"].(string)
							matchDate := regexp.MustCompile(DateRegexp)
							if matchDate.MatchString(lastDayInterface["TrDate"].(string)) {
								match, _ := strconv.ParseInt(matchDate.FindString(lastDayInterface["TrDate"].(string)), 10, 64)
								lastDay.TrDate = match
							}
							lastDay.TranNo = lastDayInterface["TranNo"].(float64)
							lastDay.CloseIndex = lastDayInterface["CloseIndex"].(float64)
							lastDay.PriorIndex = lastDayInterface["PriorIndex"].(float64)
							fmt.Println(lastDay)
						}
					}
				}
			}
			return
		}
	}
}
