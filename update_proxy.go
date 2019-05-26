package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"./model"
	"./settings"
	"./utils"
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"github.com/tebeka/selenium"
	"gopkg.in/mgo.v2/bson"
)

var (
	local_client *utils.ClientMGO
)

func main() {

	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/update_proxy.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)
	//test case
	log.Println("check to make sure it works")

	var er error
	er, local_client = utils.ConnectMGOLocalDB(utils.MongoDBInfo["docbao"])
	if er != nil {
		panic(er.Error())
	}
	defer local_client.CancelFunc()
	defer local_client.Client.Disconnect(local_client.Ctx)

	proxy_collection := local_client.Client.Database("docbao").Collection("proxy")

	// initialize value for selenium
	var webDriver selenium.WebDriver
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, er = settings.InitNewRemote(caps, utils.STANDALONESERVER); er != nil {
		fmt.Printf("Failed to open session: %s\n", er)
		return
	}
	defer webDriver.Quit()
	// client initial request on original url
	er = webDriver.Get(utils.PROXY_SOURCE)
	if er != nil {
		panic(er.Error())
	}

	// check select get 100 records proxy
	time.Sleep(500 * time.Millisecond)
	selectelm, er := webDriver.FindElement(selenium.ByCSSSelector, "select#xpp")
	if er != nil {
		log.Println("eror on get all value, ", er.Error())
		return
	}
	fmt.Println(selectelm)

	optionselm, er := selectelm.FindElements(selenium.ByTagName, "option")
	if er != nil {
		log.Println("eror on get all value, ", er.Error())
		return
	}
	fmt.Println(optionselm)
	lastOption := optionselm[2] // get 100 records
	fmt.Println(lastOption)
	lastOption.Click()
	time.Sleep(500 * time.Millisecond)

	elements, er := webDriver.FindElement(selenium.ByCSSSelector, "table tbody tr td table")
	if er != nil {
		panic("eror on get table content, " + er.Error())
	}
	body_elm, er := elements.FindElement(selenium.ByCSSSelector, "tbody")
	if er != nil {
		panic("eror on get btbody content, " + er.Error())
	}
	trs, er := body_elm.FindElements(selenium.ByCSSSelector, "tr")
	if er != nil {
		panic("eror on get tr content, " + er.Error())
	}

	for index, tr := range trs {
		if index >= 3 {
			tds, er := tr.FindElements(selenium.ByCSSSelector, "td")
			if er != nil {
				panic("eror on get td content, " + er.Error() + " css element, td")
			}

			if len(tds) > 1 {
				proxy := model.Proxy{}
				address, _ := tds[0].Text()
				pieces_addr := strings.Split(strings.Split(address, " ")[1], ":")
				schema, _ := tds[1].Text()

				if strings.Contains(strings.ToLower(schema), "https") {
					schema = "https"
				} else {
					schema = "http"
				}
				ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
				count, er := proxy_collection.Count(ctx,
					bson.M{
						"proxy_ip": pieces_addr[0],
						"port":     pieces_addr[1],
						"schema":   schema})
				if er != nil {
					log.Println("Error on count proxy, ", er.Error(), proxy)
					continue
				}
				if count == int64(0) {
					proxy.Id = primitive.NewObjectID()
					proxy.Schema = schema
					proxy.Port = pieces_addr[1]
					proxy.IP = pieces_addr[0]
					proxy.Status = true
					proxy.Created = time.Now().Format("2006-01-02 15:04:05")
					proxy.CreatedInt = time.Now().Unix()

					ctx, _ := context.WithTimeout(context.Background(), 20*time.Second)
					_, er = proxy_collection.InsertOne(ctx, &proxy)
					if er != nil {
						log.Println("Error, can not insert proxy, ", proxy)
					}
				}
			}
		}
	}
}
