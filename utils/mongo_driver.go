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

func ConnectMGOHistoryDB() (error, *ClientMGO) {
	link := "mongodb://" + USERNAME1 + ":" + PASSWORD1 + "@" + HISTORY_ONLY + ":" + PORT1 + "/" + DATABASE1
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

func ConnectMGODMPDB() (error, *ClientMGO) {

	link := "mongodb://" + USERNAME3 + ":" + PASSWORD3 + "@" + MY_HOST + ":" + PORT1 + "/" + DATABASE3
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

func ConnectMGOIPInfo() (error, *ClientMGO) {
	link := "mongodb://" + USERNAME5 + ":" + PASSWORD5 + "@" + IPINFO + ":" + PORT4 + "/" + DATABASE3
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

func ConnectMGODataOnlyDB() (error, *ClientMGO) {
	link := "mongodb://" + USERNAME2 + ":" + PASSWORD2 + "@" + ADDRESS + ":" + PORT2 + "/" + DATABASE2
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

func ConnectMGODMPLog() (error, *ClientMGO) {
	link := "mongodb://" + USERNAME4 + ":" + PASSWORD4 + "@" + DMP_LOG + ":" + PORT3 + "/" + DATABASE4
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
