package utils

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StockInfo struct {
	CatID        int    `json:"CatID" bson:"CatID"`
	StockID      int    `json:"StockID" bson:"StockID"`
	StockCode    string `json:"StockCode" bson:"StockCode"`
	ExchangeCode int    `json:"ExchangeCode" bson:"ExchangeCode"`
	ExchangeName string `json:"ExchangeName" bson:"ExchangeName"`
}

type LastDay struct {
	CloseIndex float64 `json:"CloseIndex" bson:"CloseIndex"`
	PriorIndex float64 `json:"PriorIndex" bson:"PriorIndex"`
	Change     float64 `json:"Change" bson:"Change"`
	PerChange  float64 `json:"PerChange" bson:"PerChange"`
	ChangeText string  `json:"ChangeText" bson:"ChangeText"`
	TrDate     int64   `json:"TrDate" bson:"TrDate"`
	TranNo     float64 `json:"TranNo" bson:"TranNo"`
	StockCode  string  `json:"StockCode" bson:"StockCode"`
	TrDateStr  string  `json:"TrDateStr" bson:"TrDateStr"`
}

type PriceDay struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	TradingDate  int64              `json:"TradingDate" bson:"TradingDate"`
	StockCode    string             `json:"StockCode" bson:"StockCode"`
	BasicPrice   float64            `json:"BasicPrice" bson:"BasicPrice"`
	OpenPrice    float64            `json:"OpenPrice" bson:"OpenPrice"`
	ClosePrice   float64            `json:"ClosePrice" bson:"ClosePrice"`
	HighestPrice float64            `json:"HighestPrice" bson:"HighestPrice"`
	LowestPrice  float64            `json:"LowestPrice" bson:"LowestPrice"`
	AvrPrice     float64            `json:"AvrPrice" bson:"AvrPrice"`
	Change       float64            `json:"Change" bson:"Change"`
	PerChange    float64            `json:"PerChange" bson:"PerChange"`
	ChangeText   string             `json:"ChangeText" bson:"ChangeText"`
	M_TotalVol   float64            `json:"M_TotalVol" bson:"M_TotalVol"`
	M_TotalVal   float64            `json:"M_TotalVal" bson:"M_TotalVal"`
	PT_TotalVol  float64            `json:"PT_TotalVol" bson:"PT_TotalVol"`
	PT_TotalVal  float64            `json:"PT_TotalVal" bson:"PT_TotalVal"`
	TotalVol     float64            `json:"TotalVol" bson:"TotalVol"`
	TotalVal     float64            `json:"TotalVal" bson:"TotalVal"`
	MarketCap    float64            `json:"MarketCap" bson:"MarketCap"`
	StockID      int                `json:"StockID" bson:"StockID"`
	TrID         float64            `json:"TrID" bson:"TrID"`
	TrDateStr    string             `json:"TrDateStr" bson:"TrDateStr"`
	ExchangeCode int                `json:"ExchangeCode" bson:"ExchangeCode"`
	ExchangeName string             `json:"ExchangeName" bson:"ExchangeName"`
}

type ReserveOrder struct {
	Id             primitive.ObjectID `json:"id" bson:"_id"`
	StockCode      string             `json:"StockCode" bson:"StockCode"`
	StockID        int                `json:"StockID" bson:"StockID"`
	ExchangeCode   int                `json:"ExchangeCode" bson:"ExchangeCode"`
	ExchangeName   string             `json:"ExchangeName" bson:"ExchangeName"`
	ClosePrice     float64            `json:"ClosePrice" bson:"ClosePrice"`
	TotalVol       float64            `json:"TotalVol" bson:"TotalVol"`
	TotalVal       float64            `json:"TotalVal" bson:"TotalVal"`
	BestBuy        float64            `json:"BestBuy" bson:"BestBuy"`
	BestBidVol     float64            `json:"BestBidVol" bson:"BestBidVol"`
	BestSell       float64            `json:"BestSell" bson:"BestSell"`
	BestSellVol    float64            `json:"BestSellVol" bson:"BestSellVol"`
	TotalBuyTrade  float64            `json:"TotalBuyTrade" bson:"TotalBuyTrade"`
	TotalSellTrade float64            `json:"TotalSellTrade" bson:"TotalSellTrade"`
	SpreadBSTrade  float64            `json:"SpreadBSTrade" bson:"SpreadBSTrade"`
	TotalBuyVol    float64            `json:"TotalBuyVol" bson:"TotalBuyVol"`
	TotalSellVol   float64            `json:"TotalSellVol" bson:"TotalSellVol"`
	SpreadBSVol    float64            `json:"SpreadBSVol" bson:"SpreadBSVol"`
	TradingDate    int64              `json:"TradingDate" bson:"TradingDate"`
	TradingDateStr string             `json:"TradingDateStr" bson:"TradingDateStr"`
}

type MatchOrder struct {
	Id                primitive.ObjectID `json:"id" bson:"_id"`
	TradingDate       int64              `json:"TradingDate" bson:"TradingDate"`
	StockCode         string             `json:"StockCode" bson:"StockCode"`
	ExchangeCode      int                `json:"ExchangeCode" bson:"ExchangeCode"`
	ExchangeName      string             `json:"ExchangeName" bson:"ExchangeName"`
	StockID           int                `json:"StockID" bson:"StockID"`
	TotalRoom         float64            `json:"TotalRoom" bson:"TotalRoom"`
	CurrRoom          float64            `json:"CurrRoom" bson:"CurrRoom"`
	RemainRoom        float64            `json:"RemainRoom" bson:"RemainRoom"`
	OwnedRatio        float64            `json:"OwnedRatio" bson:"OwnedRatio"`
	DiffBuySellPutVol float64            `json:"DiffBuySellPutVol" bson:"DiffBuySellPutVol"`
	BuyVol            float64            `json:"BuyVol" bson:"BuyVol"`
	PerBuyVol         float64            `json:"PerBuyVol" bson:"PerBuyVol"`
	BuyVal            float64            `json:"BuyVal" bson:"BuyVal"`
	PerBuyVal         float64            `json:"PerBuyVal" bson:"PerBuyVal"`
	SellVol           float64            `json:"SellVol" bson:"SellVol"`
	PerSellVol        float64            `json:"PerSellVol" bson:"PerSellVol"`
	SellVal           float64            `json:"SellVal" bson:"SellVal"`
	PerSellVal        float64            `json:"PerSellVal" bson:"PerSellVal"`
	DiffBuySellPutVal float64            `json:"DiffBuySellPutVal" bson:"DiffBuySellPutVal"`
	DiffBuySellVol    float64            `json:"DiffBuySellVol" bson:"DiffBuySellVol"`
	DiffBuySellVal    float64            `json:"DiffBuySellVal" bson:"DiffBuySellVal"`
	BuyPutVol         float64            `json:"BuyPutVol" bson:"BuyPutVol"`
	PerBuyPutVol      float64            `json:"PerBuyPutVol" bson:"PerBuyPutVol"`
	BuyPutVal         float64            `json:"BuyPutVal" bson:"BuyPutVal"`
	PerBuyPutVal      float64            `json:"PerBuyPutVal" bson:"PerBuyPutVal"`
	SellPutVol        float64            `json:"SellPutVol" bson:"SellPutVol"`
	PerSellPutVol     float64            `json:"PerSellPutVol" bson:"PerSellPutVol"`
	SellPutVal        float64            `json:"SellPutVal" bson:"SellPutVal"`
	PerSellPutVal     float64            `json:"PerSellPutVal" bson:"PerSellPutVal"`
	TradingDateStr    string             `json:"TradingDateStr" bson:"TradingDateStr"`
}
