package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"../settings"
)

func main() { //create your file with desired read/write permissions
	f, err := os.OpenFile("./log/province_v4_new", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//set output of logs to f
	log.SetOutput(f)

	//test case
	log.Println("check to make sure it works")
	crwl := settings.NewClient()

	url_api := "https://www.similarweb.com/"
	// url_api := "https://www.whatsmyip.org/"

	fmt.Println(url_api)

	resp, err := crwl.InitRequest(url_api)
	fmt.Println(resp, err)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err.Error())
	}
	log.Println(string(body))
	fmt.Println("string body request: ", string(body))
}
