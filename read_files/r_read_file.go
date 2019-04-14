package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	crrand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"

	"../models"
	"./structs"
	"gopkg.in/mgo.v2/bson"
)

const (
	PORT1 = "25355"
	// FOR dmp_data
	DATABASE3 = "dmp_data"
	USERNAME3 = "adreadwrite"
	PASSWORD3 = "adreadwrite!638"
	LOCALHOST = "localhost"
	MY_HOST   = "125.212.217.46"
	ADDRESS   = "125.212.217.27"
	DATABASE2 = "dmp_cookies_only_v2"
	USERNAME2 = "cookiesonly"
	PASSWORD2 = "cookiesonly!638"
	PORT2     = "41639"
)

type Message struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	customer_name := os.Args[1]

	dmp_data := ConnectDMPDB()
	defer dmp_data.Close()

	message := Message{}

	contact_online := ConnectDataOnlyDB()
	defer contact_online.Close()

	cookie_offline := contact_online.DB("dmp_cookies_only_v2").C("go_cookies_contact_offline")
	contact_info := contact_online.DB("dmp_cookies_only_v2").C("go_cookie_contact_info")
	customer_import := dmp_data.DB("dmp_data").C("customer_import")

	customerImport := models.CustomerImport{}
	message.Status = "false"
	c_er := customer_import.Find(bson.M{"status": 1, "name": customer_name}).Sort("-_id").One(&customerImport)
	if c_er != nil {
		message.Message = "error not find import from " + customer_name
		content, _ := json.Marshal(message)
		fmt.Println(string(content))

		return
	}
	path := customerImport.URL
	if strings.TrimSpace(path) == "" {
		message.Message = "error url empty"
		content, _ := json.Marshal(message)
		fmt.Println(string(content))
		return
	}
	extension := customerImport.Extension

	reader := &structs.Reader{Path: path}
	reader.SetType(extension)
	d_err := reader.GetData()
	if d_err != nil {
		message.Message = d_err.Error()
		content, _ := json.Marshal(message)
		fmt.Println(string(content))
		return
	}
	reader.GetHeaderFileImport()

	for i := 1; i < len(reader.Rows); i++ {
		// check all columns in rows not emty
		if strings.TrimSpace(strings.Join(reader.Rows[i], "")) != "" {
			customer := structs.Customer{}
			cus_err := customer.SetCustomerImport(reader.Rows[i], reader.Header)
			if cus_err != nil {
				message.Message = cus_err.Error() + ", at row: " + strconv.Itoa(i)
				content, _ := json.Marshal(message)
				fmt.Println(string(content))
				return
			}

			CookieOffline := models.GoCookieContactOffline{}
			query_string := bson.M{"$or": []bson.M{
				bson.M{"contact_info.email": Hash256(customer.Email), "user_import": customerImport.Name},
				bson.M{"contact_info.phone": Hash256(customer.Phone), "user_import": customerImport.Name}}}

			err := cookie_offline.Find(query_string).One(&CookieOffline)
			//customer exist in system
			if err == nil {
				continue
			} else {
				// initialize value for cookie_offline contact
				CookieOffline.Id = bson.NewObjectId()
				CookieOffline.Ureka = "ureka"
				b_email := []byte(customer.Email)
				b_phone := []byte(customer.Phone)
				CookieOffline.Contact_info = models.Customer{
					CustomerName: customer.CustomerName,
					Phone:        Hash256(customer.Phone),
					Email:        Hash256(customer.Email),
					Address:      customer.Address,
					Note:         customer.Note,
					PrivateEmail: Encrypt(b_email, customerImport.Name),
					PrivatePhone: Encrypt(b_phone, customerImport.Name)}
				CookieOffline.Status = 1
				CookieOffline.Core = "created_" + time.Now().Format("02-01-2006 15:04:05")
				CookieOffline.Name = ""
				CookieOffline.UserImport = customerImport.Name

				// Find cookie name exist in db
				query_string_contact := bson.M{"$or": []bson.M{
					bson.M{"contact_info.email": Hash256(customer.Email)},
					bson.M{"contact_info.phone": Hash256(customer.Phone)}}}
				CookieContact := models.ContactInfo{}
				err := contact_info.Find(query_string_contact).One(&CookieContact)
				// exist
				if err == nil {
					CookieOffline.Name = CookieContact.Name
				}

				_ = cookie_offline.Insert(CookieOffline)
			}
		}
	}
	_ = customer_import.Update(
		bson.M{"_id": customerImport.Id},
		bson.M{"$set": bson.M{"status": 3}})

	message.Status = "true"
	message.Message = "success"

	content, _ := json.Marshal(message)
	fmt.Println(string(content))
	return

}

func Hash256(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data []byte, passphrase string) string {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(crrand.Reader, nonce); err != nil {
		panic(err.Error())
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return BytesToString(ciphertext)
}

func Decrypt(data []byte, passphrase string) string {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}
	return BytesToString(plaintext)
}

func BytesToString(data []byte) string {
	return string(data)
}

func ConnectDMPDB() *mgo.Session {
	dmpHistoryDialInfo := &mgo.DialInfo{
		Addrs:    []string{MY_HOST + ":" + PORT1},
		Timeout:  1 * time.Hour,
		Database: DATABASE3,
		Username: USERNAME3,
		Password: PASSWORD3,
	}
	session, err := mgo.DialWithInfo(dmpHistoryDialInfo)
	if err != nil {
		panic(err.Error())
	}
	return session
}

func ConnectLocalDB() *mgo.Session {
	// connect local to test
	data_session, err := mgo.Dial(LOCALHOST)
	if err != nil {
		panic(err.Error())
	}
	return data_session
}

func ConnectDataOnlyDB() *mgo.Session {
	dmpHistoryDialInfo := &mgo.DialInfo{
		Addrs:    []string{ADDRESS + ":" + PORT2},
		Timeout:  1 * time.Hour,
		Database: DATABASE2,
		Username: USERNAME2,
		Password: PASSWORD2,
	}
	session, err := mgo.DialWithInfo(dmpHistoryDialInfo)
	if err != nil {
		panic(err.Error())
	}
	return session
}
