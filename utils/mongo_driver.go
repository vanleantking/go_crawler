package utils

import (
	"context"
	"fmt"
	"time"

	mongo "github.com/mongodb/mongo-go-driver/mongo"
)

type ClientMGO struct {
	Ctx        context.Context
	Client     *mongo.Client
	CancelFunc context.CancelFunc
}

type DBInfo struct {
	Address  string
	Port     string
	Database string
	Username string
	Password string
}

var MongoDBInfo = map[string]DBInfo{
	"localhost": DBInfo{
		Address:  ADDLOCALHOST,
		Port:     PORTLOCAL,
		Database: DBCK,
		Username: "",
		Password: ""},
	"docbao": DBInfo{
		Address:  ADDLOCALHOST,
		Port:     PORTLOCAL,
		Database: "docbao",
		Username: "",
		Password: ""}}

func ConnectMongoDB(dbInfo DBInfo) (error, *ClientMGO) {
	link := "mongodb://" + dbInfo.Username + ":" + dbInfo.Password + "@" + dbInfo.Address + ":" + dbInfo.Port + "/" + dbInfo.Database
	mgoclient := &ClientMGO{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mgoclient.Ctx = ctx
	mgoclient.CancelFunc = cancel
	client, err := mongo.Connect(mgoclient.Ctx, link, nil)
	mgoclient.Client = client
	if err != nil {
		fmt.Println(err, client)
		return err, mgoclient
	}

	return nil, mgoclient
}

func ConnectMGOLocalDB(dbInfo DBInfo) (error, *ClientMGO) {
	link := "mongodb://" + dbInfo.Address + ":" + dbInfo.Port + "/" + dbInfo.Database
	mgoclient := &ClientMGO{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	mgoclient.Ctx = ctx
	mgoclient.CancelFunc = cancel
	client, err := mongo.Connect(mgoclient.Ctx, link, nil)
	mgoclient.Client = client
	if err != nil {
		fmt.Println(err, client)
		return err, mgoclient
	}

	return nil, mgoclient
}
