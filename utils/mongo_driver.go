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
	"data_only": DBInfo{
		Address:  ADDRESS,
		Port:     PORT2,
		Database: DATABASE2,
		Username: USERNAME2,
		Password: PASSWORD2},
	"dmplog": DBInfo{
		Address:  DMP_LOG,
		Port:     PORT3,
		Database: DATABASE4,
		Username: USERNAME4,
		Password: PASSWORD4},
	"ip_info": DBInfo{
		Address:  IPINFO,
		Port:     PORT4,
		Database: DATABASE3,
		Username: USERNAME5,
		Password: PASSWORD5},
	"dmp_data": DBInfo{
		Address:  MY_HOST,
		Port:     PORT1,
		Database: DATABASE3,
		Username: USERNAME3,
		Password: PASSWORD3},
	"history_only": DBInfo{
		Address:  HISTORY_ONLY,
		Port:     PORT1,
		Database: DATABASE1,
		Username: USERNAME1,
		Password: PASSWORD1},
	"localhost": DBInfo{
		Address:  ADDLOCALHOST,
		Port:     PORTLOCAL,
		Database: DBCK,
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
