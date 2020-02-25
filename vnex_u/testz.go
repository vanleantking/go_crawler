package main

import (
	"context"
	"fmt"
	"time"

	"../utils"
	structs "./utils"
	"github.com/mongodb/mongo-go-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

type Cmts struct {
	Comments []structs.DetailComment `json:"comments" bson:"comments"`
}

var (
	ipClient *utils.ClientMGO
)

func main() {
	var er error
	er, ipClient = utils.ConnectMongoDB(utils.MongoDBInfo["ip_info"])
	if er != nil {
		panic(er.Error())
	}
	defer ipClient.CancelFunc()
	defer ipClient.Client.Disconnect(ipClient.Ctx)

	vnexLinksC := ipClient.Client.Database("dmp_data").Collection("vnexpress_links")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var cmts Cmts
	er = vnexLinksC.FindOne(
		ctx,
		bson.M{"link": "https://vnexpress.net/tam-su/toi-hoi-han-khi-lay-chong-chenh-hoc-van-gia-canh-3979855.html"},
		options.FindOne().SetProjection(
			bson.M{"comments": 1})).Decode(&cmts)

	if er != nil {
		panic(er.Error())
	}

	zz := structs.GetUniqueDetailCmt(cmts.Comments)
	fmt.Println(zz)
}
