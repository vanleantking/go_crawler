package utils

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	// "fmt"

	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var remove_list = map[string]bool{
	"Trang chủ": true,
	"trang chu": true,
	"diễn đàn":  true,
	"dien dan":  true}

var MaxMsTimeOut = 500
var MinMsTimeOut = 200
var TimeOutRange = []int{200, 300, 400, 500, 500, 700, 800, 900, 1000, 1100, 1200, 1300, 1400, 1500}
var TimeOutWideRange = []int{880, 1000, 1120, 1450, 1890, 2030, 2340, 2500, 2590, 2610, 2790, 2900, 3000, 3500, 4000, 5000}

var RegexRemoves = []string{
	`\/\*<!\[CDATA\[.*?\]\]>\*\/`,
	`\(?function\s*\(\s*\)\s*{.*?}\)?\(?\)?;?`,
	`<img([\w\W]+?)\/>`,
	`<[^>]*>`}

// replace multi space with 1 space
func StrimSpace(str string) string {
	str = strings.TrimSpace(str)
	if str == "" {
		return ""
	}
	re := regexp.MustCompile(`\s+`)
	str = re.ReplaceAllLiteralString(str, " ")
	return str
}

func GetCrawlURL(log_url string) (string, error) {
	parse_url, err := url.Parse(log_url)
	if err != nil {
		return "", err
	}
	host := parse_url.Host
	host_component := strings.Split(host, ".")

	result_url := log_url
	index := 0

	for _, value := range host_component {
		// remove f. in url
		if value == "f" {
			index = strings.Index(log_url, "f.")
			result_url = log_url[:index] + log_url[index+2:]
		}

		//remove m. in url
		if value == "m" {
			index = strings.Index(log_url, "m.")
			result_url = log_url[:index] + log_url[index+2:]
		}
	}

	return result_url, nil
}

// Get content news from list of class split by "|" charater from DB
func GetContentFromTag(tag string, doc *goquery.Document) string {
	content := doc.Find(tag).Text()

	return content
}

// Get content news from list of class split by "|" charater from DB
func GetContentFromClass(classes string, doc *goquery.Document) string {
	piece_classes := strings.Split(classes, "|")
	content := ""
	for _, class := range piece_classes {
		contentSelection := doc.Find(class)

		contentSelection.Find("script").Remove()
		contentSelection.Find("style").Remove()
		content = contentSelection.Text()
		if content != "" {
			break
		}
	}
	content = CleanDataContent(content)
	return content
}

func GetContentVGTFromClass(classes string, doc *goquery.Document) string {
	piece_classes := strings.Split(classes, "|")
	content := ""
	for _, class := range piece_classes {
		contentSelection := doc.Find(class)

		contentSelection.Find("script").Remove()
		contentSelection.Find("style").Remove()

		content = contentSelection.Text()
		if strings.TrimSpace(content) != "" {
			break
		}
	}
	content = CleanDataContent(content)
	return content
}

// Get category news from list of class split by "|" charater from DB
func GetCategoryFromClass(classes string, doc *goquery.Document) string {
	piece_classes := strings.Split(classes, "|")
	content := ""
	for _, class := range piece_classes {
		doc.Find(class).Each(func(i int, piece *goquery.Selection) {
			// process for category (list li - ul), remove text Trang chu, dien dan in category
			piece_str := strings.ToLower(strings.TrimSpace(piece.Text()))
			if _, ok := remove_list[piece_str]; !ok {
				// Add special charater "|" among category
				content = piece_str
			}
		})
		if content != "" {
			break
		}
	}
	return content
}

// generate random time from [200, 1500] ms
func RandInRange() time.Duration {
	return time.Duration(rand.Intn(len(TimeOutRange))) * time.Microsecond
}

// generate random time from [880, 5000] ms
func RangeWideTimeOut() time.Duration {
	return time.Duration(rand.Intn(len(TimeOutWideRange))) * time.Microsecond
}

func removeCmt(content string) string {
	re, err := regexp.Compile(`\/\*<!\[CDATA\[.*?\]\]>\*\/`)
	if err != nil {
		return content
	}
	result := re.ReplaceAllLiteralString(content, " ")
	return result
}

func removeFunctionScript(content string) string {
	re, err := regexp.Compile(`\(?function\s*\(\s*\)\s*{.*?}\)?\(?\)?;?`)
	if err != nil {
		return content
	}
	result := re.ReplaceAllLiteralString(content, " ")
	return result
}

func removeTagHTML(content string) string {
	re1, err := regexp.Compile(`<img([\w\W]+?)\/>`)
	if err != nil {
		return content
	}
	result := re1.ReplaceAllLiteralString(content, " ")

	re2, err := regexp.Compile(`<[^>]*>`)
	if err != nil {
		return content
	}
	result = re2.ReplaceAllLiteralString(content, " ")
	return result
}

func CleanDataContent(content string) string {
	result := strings.TrimSpace(content)
	if result == "" {
		return content
	}

	result = removeCmt(result)
	result = removeFunctionScript(result)
	result = removeTagHTML(result)
	return result
}

func ConnectMySQL() (*sql.DB, error) {
	MyDb, err := sql.Open("mysql", MY_USERNAME+":"+MY_PASSWORD+"@"+"tcp("+MY_HOST+":3306)"+"/"+MY_DATABASE+"?parseTime=true")
	return MyDb, err
}

func GetDomainName(hostname string) string {
	return strings.Replace(hostname, "www.", "", -1)
}

func GetCategoryLink(list_news string, title_news string, doc *goquery.Document, host_name string, domain string) []string {
	var links = []string{}
	doc.Find(list_news).Each(func(i int, s *goquery.Selection) {
		href := s.Find(title_news)
		if link, ok := href.Attr("href"); ok {
			if !strings.Contains(link, domain) {
				link = host_name + link
			}
			links = append(links, link)
		}
	})
	var result = make([]string, len(links))
	copy(result, links)
	return result
}
