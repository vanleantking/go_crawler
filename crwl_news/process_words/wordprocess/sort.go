package wordprocess

import "gopkg.in/mgo.v2/bson"

type Pair struct {
	Key   string  `json:"key"  bson:"key"`
	Value float32 `json:"value"  bson:"value"`
}

type WordCount struct {
	Id           bson.ObjectId `json:"id" bson:"_id"`
	CreatedAt    string        `json:"createdAt" bson:"createdAt"`
	PairList     `json:"keywords" bson:"keywords"`
	UpdatedAt    string              `json:"updatedAt" bson:"updatedAt"`
	Deleted      bool                `json:"delete" bson:"delete"`
	URL          string              `json:"url" bson:"url"`
	Domain       string              `json:"domain" bson:"domain"`
	Title        string              `json:"title" bson:"title"`
	Category     string              `json:"category" bson:"category"`
	WeekNumber   int64               `json:"week_number" bson:"week_number"`
	RandNumber   int                 `json:"rand_number" bson:"rand_number"`
	Language     string              `json:"language" bson:"language"`
	Class        string              `json:"class" bson:"class"`
	Content      string              `json:"content" bson:"content"`
	CookiesRead  []map[string]string `json:"cookies_read" bson:"cookies_read"`
	CategoryNews string              `json:"category_news" bson:"category_news"`
	Status       int                 `json:"status" bson:"status"`
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
