package model

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
)

type Proxy struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Port       string             `json:"port" bson:"port"`
	IP         string             `json:"proxy_ip" bson:"proxy_ip"`
	Schema     string             `json:"schema" bson:"schema"`
	Status     bool               `json:"status" bson:"status"`
	Created    string             `json:"created" bson:"created"`
	CreatedInt int64              `json:"created_int" bson:"created_int"`
}
