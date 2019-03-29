package main

// implement selenium crawl to get data from
import (
	"fmt"
	"log"
	"strings"
	"time"

	"./settings"
	"./utils"
	"github.com/tebeka/selenium"
)

var (
	San = map[string][]string{
		"HOSE": []string{"AA", "AAM"}}
	StartDate        = ""
	EndDate          = ""
	IDStatisticPrice = "#statistic-price .table tbody"
	IDTabView        = "view-tab"
	Result           = map[string][]string{}
)

func main() {
	// initialize value for selenium
	var webDriver selenium.WebDriver
	var err error
	caps := settings.SetChomeCapabilities()

	// connect to selenium Standalone alone (run on java jar package)
	if webDriver, err = settings.InitNewRemote(caps, utils.STANDALONESERVER); err != nil {
		fmt.Printf("Failed to open session: %s\n", err)
		log.Println("Failed to open session: " + err.Error())
		return
	}
	defer webDriver.Quit()

	//Process on original log_url
	log_url := strings.TrimSpace("https://finance.vietstock.vn/ket-qua-giao-dich")

	var wcontent, wtitle string

	// client initial request on original url
	err = webDriver.Get(log_url)
	// error on log url
	if err != nil {
		log.Println("Init request error", err.Error())

		fmt.Println("Erorrrrrrrrrrrrrrrrrrrrrrrrrr", err.Error())
	} else {
		// sleep for a while for fully loaded javascript
		time.Sleep(200 * time.Millisecond)
		// get title
		if title, err := webDriver.Title(); err == nil {
			wtitle = title
		} else {
			log.Println("Failed to get page title: " + " " + err.Error())
		}

		wcontent, err = utils.SeleniumGetContentFromClass(IDStatisticPrice, webDriver)

		fmt.Println("titleeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee", wcontent, wtitle, err)
	}
}
