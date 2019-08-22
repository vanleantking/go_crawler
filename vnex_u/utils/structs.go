package utils

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type LinkCrawler struct {
	ID           primitive.ObjectID `json:"id" bson:"_id"`
	Link         string             `json:"link" bson:"link"`
	Title        string             `json:"title" bson:"title"`
	Comments     []DetailComment    `json:"comments" bson:"comments"`
	Created      int64              `json:"created" bson:"created"`
	TotalComment int                `json:"total_comment" bson:"total_comment"`
	Status       int                `json:"status" bson:"status"`
}

type DetailComment struct {
	ProfileLink string `json:"profile_link" bson:"profile_link"`
	UserName    string `json:"user_name" bson:"user_name"`
	Content     string `json:"content" bson:"content"`
}

type VNExUser struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	ProfileLink string             `json:"profile_link" bson:"profile_link"`
	UserID      string             `json:"user_id" bson:"user_id"`
	UserName    string             `json:"user_name" bson:"user_name"`
	JoinedDate  string             `json:"joined_date" bson:"joined_date"`
	Created     int64              `json:"created" bson:"created"`
	Status      int                `json:"status" bson:"status"`
}
