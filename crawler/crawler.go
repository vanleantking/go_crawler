package crawler

// implement crawler web data from existing config
import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"../settings"
	"../utils"
)

var ConfigWeb map[string]settings.WebsiteConfig
var (
	BreakTime = [10]int64{700, 600, 500, 1000, 10000, 5000, 1500, 800, 3000, 2500}
)

func init() {
	ConfigWeb = settings.SetConfig()
}

type Crawler struct {
	Client *settings.Client
	WS     map[string]settings.WebsiteConfig
}

type Crresult struct {
	content       string
	title         string
	category_news string
	keyword       []string
	description   string
	meta          map[string]string
	publish_date  string
}

type Result struct {
	Content      string
	Title        string
	CategoryNews string
	Keyword      []string
	Description  string
	Meta         map[string]string
	PublishDate  string
}

func InitCrawler() *Crawler {
	return &Crawler{
		WS:     ConfigWeb,
		Client: settings.NewClient()}

}

func (crw *Crawler) NewClient() {
	crw.Client = settings.NewClient()
}

func (crw *Crawler) CrawlerURL(log_url string) (*Result, error) {

	var res *http.Response
	var err error

	referer, _ := getHostFromURL(log_url)
	ws := crw.getWebConfig(log_url)
	// client initial request on original url
	if ws.SpecialHeader {
		res, err = crw.Client.InitRequest2(log_url, referer, ws.Domain)
	} else {
		res, err = crw.Client.InitRequest(log_url)
	}
	re := &Result{}

	if err != nil {
		return re, err
	}
	defer res.Body.Close()

	var result = Crresult{}

	// Continue if Response code is success
	if res.StatusCode != 200 {
		msg := "status code error: " + " " + strconv.Itoa(res.StatusCode) + " " + res.Status
		return re, errors.New(msg)
	}
	// Load the HTML document
	doc, docer := goquery.NewDocumentFromReader(res.Body)
	if docer != nil {
		return re, docer
	}

	// Find the content page
	result.content = utils.GetContentFromClass(ws.ContentStruct, doc)
	result.content = utils.StrimSpace(result.content)

	result.category_news = utils.GetCategoryFromClass(ws.CategoryStruct, doc)
	result.category_news = utils.StrimSpace(result.category_news)

	// find title page
	result.title = utils.GetContentFromTag("title", doc)
	result.title = utils.StrimSpace(result.title)

	// Can get content from original class structure and log url
	if result.content == "" {
		parsed_url, err := utils.GetCrawlURL(log_url)

		// Parse processed url success
		if err == nil {
			// initial another request on parse url
			if ws.SpecialHeader {
				res, err = crw.Client.InitRequest2(parsed_url, referer, ws.Domain)
			} else {
				res, err = crw.Client.InitRequest(parsed_url)
			}

			// initial request error
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()

			// Continue if Response code is success, process on url parsed
			if res.StatusCode == 200 {
				// Load the HTML document
				doc, docer := goquery.NewDocumentFromReader(res.Body)
				// continue to next content_news when can not read, and do nothing
				if docer != nil {
					return nil, docer
				}

				// Find the content
				result.content = utils.GetContentFromClass(ws.ContentStruct, doc)
				result.content = utils.StrimSpace(result.content)

				// Find the category on new url
				result.category_news = utils.GetCategoryFromClass(ws.CategoryStruct, doc)
				result.category_news = utils.StrimSpace(result.category_news)
			} else {
				msg := "status code error: " + " " + strconv.Itoa(res.StatusCode)
				return re, errors.New(msg)
			}
			// Parse url failed
		} else {
			msg := "parse URL failed: " + ws.Url
			return re, errors.New(msg)
		}
	}
	result.GetKeywords(doc, ws.Keywords)
	result.GetDescription(doc, ws.Description)
	result.GetMetaTags(doc)
	result.GetPublishDate(doc, ws.PublishDate)
	re = &Result{
		Content:      result.content,
		Title:        result.title,
		CategoryNews: result.category_news,
		Keyword:      result.keyword,
		Description:  result.description,
		Meta:         result.meta,
		PublishDate:  result.publish_date}
	return re, nil
}

func (crw *Crawler) getWebConfig(log_url string) settings.WebsiteConfig {
	u, err := url.Parse(log_url)
	if err != nil {
		panic(err)
	}
	domain := utils.GetDomainName(u.Hostname())
	return crw.WS[domain]
}

func (result *Crresult) GetMetaTag(tag string, doc *goquery.Document) string {
	metaContent := ""

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); name == tag {
			metaContent, _ = s.Attr("content")
		}
	})
	return metaContent
}

func (result *Crresult) GetPublishDate(doc *goquery.Document, publish_class string) {
	contentSelection := doc.Find(publish_class)
	result.publish_date = contentSelection.Text()
}

func (result *Crresult) GetMetaTags(doc *goquery.Document) {
	var metas = map[string]string{}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		metas[name], _ = s.Attr("content")
	})
	result.meta = metas
}

func (result *Crresult) GetKeywords(doc *goquery.Document, keyword_class string) {
	keywordstr := result.GetMetaTag(keyword_class, doc)
	var keywords = []string{}
	pieces := strings.Split(keywordstr, ",")
	for _, k := range pieces {
		keywords = append(keywords, strings.TrimSpace(k))
	}
	result.keyword = keywords
}

func (result *Crresult) GetDescription(doc *goquery.Document, description_class string) {
	result.description = result.GetMetaTag(description_class, doc)
}

// request for auto get link from category or hompage on web config
func (crw *Crawler) FetchURL() []string {
	var results = []string{}
	var links = []string{}
	crawl_url := ""
	for domain, config := range crw.WS {
		if config.PaginateRegex != "" {
			for i := 1; i <= 10; i++ {
				// break 5s before crawl next page
				time.Sleep(10 * time.Second)
				crawl_url = config.Url
				if i > 1 {
					crawl_url = config.Url + config.PaginateRegex + strconv.Itoa(i)
				}

				links = crw.crawlSingleLink(crawl_url, domain)
				results = append(results, links...)
			}
		} else {
			crawl_url = config.Url
			links = crw.crawlSingleLink(crawl_url, domain)
			results = append(results, links...)
		}
	}
	return results
}

// request for auto get link from category or hompage on web config
func (crw *Crawler) FetchSingleURL(domain string, config settings.WebsiteConfig) []string {
	var results = []string{}
	var links = []string{}
	crawl_url := ""

	fmt.Println("config, ", config, domain, config.PaginateRegex)
	if config.PaginateRegex != "" {
		for i := 1; i <= 10; i++ {
			// break random time before crawl next page
			break_time := rand.Intn(len(BreakTime))
			duration := time.Duration(BreakTime[break_time]) * time.Microsecond
			time.Sleep(duration)
			crawl_url = config.Url
			if i > 1 {
				crawl_url = config.Url + config.PaginateRegex + strconv.Itoa(i)
			}

			links = crw.crawlSingleLink(crawl_url, domain)
			results = append(results, links...)
		}
	} else {
		crawl_url = config.Url
		links = crw.crawlSingleLink(crawl_url, domain)
		results = append(results, links...)
	}
	return results
}

func (crw *Crawler) crawlSingleLink(crawl_link string, domain string) []string {
	var res *http.Response
	var err error

	// client initial request on original url
	referer, _ := getHostFromURL(crawl_link)
	ws := crw.WS[domain]
	if ws.SpecialHeader {
		res, err = crw.Client.InitRequest2(crawl_link, referer, domain)
	} else {
		res, err = crw.Client.InitRequest(crawl_link)
	}

	if err != nil {
		log.Println(err.Error(), crawl_link)
		return make([]string, 0)
	}
	defer res.Body.Close()

	// Continue if Response code is success
	if res.StatusCode != 200 {
		msg := "status code error: " + " " + res.Status + " " + crawl_link
		log.Println(msg)
		return make([]string, 0)
	}
	// Load the HTML document
	doc, docer := goquery.NewDocumentFromReader(res.Body)
	if docer != nil {
		log.Println("Error, ", docer.Error())
		return make([]string, 0)
	}
	return utils.GetCategoryLink(ws.ListNews, ws.TitleNews, doc, referer, domain)
}

func getHostFromURL(url_str string) (string, error) {
	// Parse the URL and ensure there are no errors.
	u, err := url.Parse(url_str)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}

func (crw *Crawler) FetchURLs(crawl_link string) []string {
	var results = []string{}
	var links = []string{}
	links = crw.crawlSingleLink(crawl_link, "vnexpress.net")
	results = append(results, links...)
	return results
}
