package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type News struct {
	Id           primitive.ObjectID `json:"id" bson:"_id"`
	URL          string             `json:"url" bson:"url"`
	Domain       string             `json:"domain" bson:"domain"`
	Title        string             `json:"title" bson:"title"`
	Category     string             `json:"category" bson:"category"`           // ([cat1], [cat2])
	CategoryType string             `json:"category_type" bson:"category_type"` // ([cat1], [cat2])
	Content      string             `json:"content" bson:"content"`
	Description  string             `json:"description" bson:"description"`
	Keywords     string             `json:"keywords" bson:"keywords"` // ([key1], [key2])
	Keyword      []string           `json:"keyword" bson:"keyword"`   // ([key1], [key2])
	NewKeyWords  string             `json:"new_keywords" bson:"new_keywords"`
	Meta         string             `json:"meta" bson:"meta"`
	PublishDate  string             `json:"publish_date" bson:"publish_date"`
	Words        string             `json:"words" bson:"words"`
	Status       int                `json:"status" bson:"status"`
	DateTime     string             `json:"date_time" bson:"date_time"`
	CreatedInt   int64              `json:"created_int" bson:"created_int"`
	UpdatedStr   string             `json:"updated_str" bson:"updated_str"`
}
