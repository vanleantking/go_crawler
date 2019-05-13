package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"../models"
	"../utils"
	"./structs"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	//create your file with desired read/write permissions
	f, err := os.OpenFile("./log/read_file.txt", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}

	//defer to close when you're done with it, not because you think it's idiomatic!
	defer f.Close()

	//set output of logs to f
	log.SetOutput(f)

	//test case
	log.Println("check to make sure it works")

	dmp_data := utils.ConnectDMPDB()
	defer dmp_data.Close()

	contact_offline := utils.ConnectLocalDB()
	defer contact_offline.Close()

	// contact_online := utils.ConnectDataOnlyDB()
	// defer contact_online.Close()

	cookie_offline := contact_offline.DB("dmp_cookies_only_v2").C("go_cookie_contact_offline")
	// contact_info := contact_online.DB("dmp_cookies_only_v2").C("go_cookie_contact_info")
	customer_import := dmp_data.DB("dmp_data").C("customer_import")

	page := 1
	condition := true
	for condition {
		if page == 2 {
			break
		}
		customerImports := customer_import.Find(bson.M{"status": 1}).Sort("-_id").Skip((page - 1) * 1).Limit(1).Iter()
		customerImport := models.CustomerImport{}
		condition = false

		for customerImports.Next(&customerImport) {
			condition = true
			path := "https://docs.google.com/spreadsheets/d/1xdTHMtIGr-xy8Bwi0qTXaEzoytBIYF0vCkkz2CZzPDA/edit#gid=0"
			if strings.TrimSpace(path) == "" {
				log.Println("url is empty: ", customerImport.Id)
				continue
			}
			extension := "google_sheet" //customerImport.Extension

			reader := &structs.Reader{Path: path}
			reader.SetType(extension)
			reader.GetData()
			reader.GetHeaderFileImport()

			for i := 1; i < len(reader.Rows); i++ {
				// check all columns in rows not emty
				if strings.TrimSpace(strings.Join(reader.Rows[i], "")) != "" {
					customer := structs.Customer{}
					customer.SetCustomerImport(reader.Rows[i], reader.Header)

					CookieOffline := models.GoCookieContactOffline{}
					query_string := bson.M{"$or": []bson.M{
						bson.M{"contact_info.email": utils.Hash256(customer.Email), "user_import": customerImport.Name},
						bson.M{"contact_info.phone": utils.Hash256(customer.Phone), "user_import": customerImport.Name}}}

					err := cookie_offline.Find(query_string).One(&CookieOffline)
					//customer exist in system
					if err == nil {
						log.Println("customer exist: ", customer.Email)
						log.Println("customer descrypt email: ", string(utils.Decrypt([]byte(CookieOffline.Contact_info.PrivateEmail), customerImport.Name)))
						continue
					} else {
						// initialize value for cookie_offline contact
						CookieOffline.Id = bson.NewObjectId()
						CookieOffline.Ureka = "ureka"
						b_email := []byte(customer.Email)
						b_phone := []byte(customer.Phone)
						CookieOffline.Contact_info = models.Customer{
							CustomerName: customer.CustomerName,
							Phone:        utils.Hash256(customer.Phone),
							Email:        utils.Hash256(customer.Email),
							Address:      customer.Address,
							Note:         customer.Note,
							PrivateEmail: utils.Encrypt(b_email, customerImport.Name),
							PrivatePhone: utils.Encrypt(b_phone, customerImport.Name)}
						CookieOffline.Status = 1
						CookieOffline.Core = "created_" + time.Now().Format("02-01-2006 15:04:05")
						CookieOffline.Name = ""
						CookieOffline.UserImport = customerImport.Name

						// Find cookie name exist in db
						// CookieContact := models.ContactInfo{}
						// err := contact_info.Find(query_string).One(&CookieContact)
						// if err == nil {
						// 	CookieOffline.Name = CookieContact.Name
						// }

						err = cookie_offline.Insert(CookieOffline)
						if err != nil {
							log.Println("Can not insert customer: ", customer.Email)
							panic(err.Error())
						} else {
							log.Println("insert customer success: ", customer.Email)
						}
					}
				}
			}
		}
		page++
	}

	fmt.Println("success")

}
