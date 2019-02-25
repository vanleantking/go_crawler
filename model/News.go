package model

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type News struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	URL         string             `json:"url" bson:"url"`
	Category    string             `json:"category" bson:"category"`
	Content     string             `json:"content" bson:"content"`
	Description string             `json:"description" bson:"description"`
	Keywords    string             `json:"keywords" bson:"keywords"`
	Meta        string             `json:"meta" bson:"meta"`
	PublishDate string             `json:"publish_date" bson:"publish_date"`
	Words       string             `json:"words" bson:"words"`
}
