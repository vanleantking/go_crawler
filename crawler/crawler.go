package crawler

// implement crawler web data from existing config
import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"../settings"
	"../utils"
)

var ConfigWeb = settings.SetConfig()

type Crawler struct {
	Client *settings.Client
	WS     *settings.WebsiteConfig
	*Result
}

type Result struct {
	content       string
	title         string
	category_news string
	keyword       []string
	description   string
	meta          map[string]string
}

func (crw *Crawler) Getresult() (string, string, string, string, []string, map[string]string) {
	return crw.title, crw.content, crw.category_news, crw.description, crw.keyword, crw.meta
}

func (crw *Crawler) NewClient() {
	client := settings.NewClient()
	crw.Client = client
}

func (crw *Crawler) CrawlerURL(log_url string) error {

	crw.settingWebConfig(log_url)

	var res *http.Response
	var err error

	// client initial request on original url
	if crw.WS.SpecialHeader {
		referer, _ := getHostFromURL(log_url)
		res, err = crw.Client.InitRequest2(log_url, referer)
	} else {
		res, err = crw.Client.InitRequest(log_url)
	}

	if err != nil {
		return err
	}
	defer res.Body.Close()

	var result = Result{}

	// Continue if Response code is success
	if res.StatusCode != 200 {
		msg := "status code error: " + " " + strconv.Itoa(res.StatusCode)
		return errors.New(msg)
	}
	// Load the HTML document
	doc, docer := goquery.NewDocumentFromReader(res.Body)
	if docer != nil {
		return docer
	}

	// Find the content page
	result.content = utils.GetContentFromClass(crw.WS.ContentStruct, doc)
	result.content = utils.StrimSpace(result.content)

	result.category_news = utils.GetCategoryFromClass(crw.WS.CategoryStruct, doc)
	result.category_news = utils.StrimSpace(result.category_news)

	// find title page
	result.title = utils.GetContentFromTag("title", doc)
	result.title = utils.StrimSpace(result.title)

	// Can get content from original class structure and log url
	if result.content == "" {
		parsed_url, err := utils.GetCrawlURL(log_url)

		// Parse processed url success
		if err == nil {
			// only initial another request if url is parsed
			if parsed_url != log_url {

				// initial another request on parse url
				res, err := crw.Client.InitRequest(parsed_url)

				// initial request error
				if err != nil {
					return err
				}
				defer res.Body.Close()

				// Continue if Response code is success, process on url parsed
				if res.StatusCode == 200 {
					// Load the HTML document
					doc, docer := goquery.NewDocumentFromReader(res.Body)
					// continue to next content_news when can not read, and do nothing
					if docer != nil {
						return docer
					}

					// Find the content
					result.content = utils.GetContentFromClass(crw.WS.ContentStruct, doc)
					result.content = utils.StrimSpace(result.content)

					// Find the category on new url
					result.category_news = utils.GetCategoryFromClass(crw.WS.CategoryStruct, doc)
					result.category_news = utils.StrimSpace(result.category_news)
				}
			}

			// Parse url failed
		} else {
			msg := "parse URL failed: " + crw.WS.Url
			return errors.New(msg)
		}
	}
	crw.Result = &result
	crw.GetKeywords(doc)
	crw.GetDescription(doc)
	crw.GetMetaTags(doc)
	return nil
}

func (crw *Crawler) settingWebConfig(log_url string) {
	u, err := url.Parse(log_url)
	if err != nil {
		panic(err)
	}
	domain := utils.GetDomainName(u.Hostname())
	websiteStruct := ConfigWeb[domain]
	crw.WS = &websiteStruct
}

func (result *Result) GetMetaTag(tag string, doc *goquery.Document) string {
	metaContent := ""

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if name, _ := s.Attr("name"); name == tag {
			metaContent, _ = s.Attr("content")
		}
	})
	return metaContent
}

func (crw *Crawler) GetMetaTags(doc *goquery.Document) {
	var metas = map[string]string{}

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		metas[name], _ = s.Attr("content")
	})
	crw.meta = metas
}

func (crw *Crawler) GetKeywords(doc *goquery.Document) {
	keywordstr := crw.Result.GetMetaTag(crw.WS.Keywords, doc)
	var keywords = []string{}
	pieces := strings.Split(keywordstr, ",")
	for _, k := range pieces {
		keywords = append(keywords, strings.TrimSpace(k))
	}
	crw.Result.keyword = keywords
}

func (crw *Crawler) GetDescription(doc *goquery.Document) {
	crw.Result.description = crw.Result.GetMetaTag(crw.WS.Description, doc)
}

func (crw *Crawler) FetchURL() {
	for _, config := range ConfigWeb {
		// for i := 0; i < 2; i++ {
		crawl_url := config.Url // + config.PaginateRegex + strconv.Itoa(i)
		var res *http.Response
		var err error

		// client initial request on original url
		if config.SpecialHeader {
			referer, _ := getHostFromURL(crawl_url)
			res, err = crw.Client.InitRequest2(crawl_url, referer)
		} else {
			res, err = crw.Client.InitRequest(crawl_url)
		}

		if err != nil {
			panic(err.Error())
		}
		defer res.Body.Close()

		// Continue if Response code is success
		if res.StatusCode != 200 {
			msg := "status code error: " + " " + strconv.Itoa(res.StatusCode)
			// return errors.New(msg)
			panic(msg)
		}
		// Load the HTML document
		doc, docer := goquery.NewDocumentFromReader(res.Body)
		if docer != nil {
			panic(docer.Error())
		}
		links := utils.GetCategoryLink(config.ListNews, config.TitleNews, doc)
		fmt.Println(links)
		// }
	}
}

func getHostFromURL(url_str string) (string, error) {
	// Parse the URL and ensure there are no errors.
	u, err := url.Parse(url_str)
	if err != nil {
		return "", err
	}
	return u.Scheme + "://" + u.Host, nil
}
