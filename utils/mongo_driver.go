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

func ConnectMongoDB(database string) (error, *ClientMGO) {
	dbInfo := DBCONFIG[database]
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

func ConnectMGOLocalDB() (error, *ClientMGO) {
	databasename := "dmp_cookies_ony_v2"
	port := "27017"
	link := "mongodb://" + LOCALHOST + ":" + port + "/" + databasename
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
