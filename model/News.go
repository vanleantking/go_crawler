package model

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type News struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	URL         string             `json:"url" bson:"url"`
	Category    string             `json:"category" bson:"category"` // ([cat1], [cat2])
	Content     string             `json:"content" bson:"content"`
	Description string             `json:"description" bson:"description"`
	Keywords    string             `json:"keywords" bson:"keywords"` // ([key1], [key2])
	NewKeyWords string             `json:"new_keywords" bson:"new_keywords"`
	Meta        string             `json:"meta" bson:"meta"`
	PublishDate string             `json:"publish_date" bson:"publish_date"`
	Words       string             `json:"words" bson:"words"`
	Status      bool               `json:"status" bson:"status"`
	CreatedInt  int64              `json:"created_int" bson:"created_int"`
	UpdatedInt  int64              `json:"updated_int" bson:"updated_int"`
}
