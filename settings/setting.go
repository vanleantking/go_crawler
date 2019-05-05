package settings

/*
Implement go-query to crawl proxy content
*/

import (
	"crypto/tls"
	"errors"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

type UserAgent struct {
	Name string
}

type Proxy struct {
	ProxyIP string `json:"proxy_ip"`
	Port    string `json:"port"`
	Schema  string `json:"schema"`
}

type Proxies struct {
	ActiveProxy []Proxy `json:"active_proxies"`
}

type Response struct {
	*http.Response
}

// Custom new request type for setting header on http.Request
type CrRequest struct {
	*http.Request
}

func (cr *CrRequest) SetHeader(header Header) {
	cr.Request.Header.Add("method", header.Method)
	cr.Request.Header.Add("user-agent", header.UserAgent)
	cr.Request.Header.Add("referer", header.Referrer)
	cr.Request.Header.Add("accept", header.Accept)
	cr.Request.Header.Add("accept-language", header.AcceptLanguage)
	cr.Request.Header.Add("cache-control", header.CacheControl)
	// cr.Request.Header.Add("content-type", header.ContentType)
	cr.Request.Header.Add("upgrade-insecure-requests", header.UpdateInsecCureRequest)
	cr.Request.Header.Add("connection", header.Connection)
	if header.Host != "" {
		cr.Request.Header.Add("host", header.Host)
	}
}

type Header struct {
	UserAgent              string
	Referrer               string
	Accept                 string
	AcceptEncoding         string
	AcceptLanguage         string
	Pragma                 string
	Method                 string
	Scheme                 string
	CacheControl           string
	ContentType            string
	UpdateInsecCureRequest string
	Connection             string
	Host                   string
}

func (header *Header) SetUserAgent(ua string) {
	header.UserAgent = ua
}

type Client struct {
	client *http.Client
}

func (client *Client) NewRequest(method, urlStr string, body io.Reader) (*CrRequest, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	// defer req.Body.Close()

	cr_request := &CrRequest{Request: req}

	return cr_request, nil
}

func (client *Client) SetProxy() {

	condition := true
	for condition {
		var proxyUrl *url.URL
		proxy_index := rand.Intn(len(ProxyList.ActiveProxy))
		proxy := ProxyList.ActiveProxy[proxy_index]
		proxyUrl, err := url.Parse(proxy.Schema + "://" + proxy.ProxyIP + ":" + proxy.Port)
		if err != nil {
			condition = true
		} else {
			client.client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl), TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
			condition = false
		}
	}
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// retrieve another proxy 3 time request when failed
	flag := 0
	re_error := errors.New("")
	for flag < 30 {
		resp, err := c.client.Do(req)
		if err != nil {
			flag++
			re_error = err
			c.SetProxy()
		} else {
			return resp, err
		}
	}

	return nil, errors.New("--------------------------Refused connection after 10 time retrieve----------------------------------------" + re_error.Error())
}

func (client *Client) InitRequest(url string) (*http.Response, error) {
	header := Header{
		Referrer:               "https://www.google.com.vn/",
		Accept:                 "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		AcceptLanguage:         "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		Pragma:                 "no-cache",
		Method:                 "GET",
		ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: "1",
		CacheControl:           "max-age=0",
		Connection:             "keep-alive"}

	header.SetUserAgent(UserAgents[rand.Intn(len(UserAgents))])

	cr_request, er := client.NewRequest(header.Method, url, nil)

	if er != nil {
		panic(er.Error())
	}
	cr_request.SetHeader(header)

	return client.Do(cr_request.Request, nil)
}

func (client *Client) InitRequest2(url string, host_name string, domain string) (*http.Response, error) {
	header := Header{
		Referrer:               host_name,
		Accept:                 "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		AcceptLanguage:         "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		Pragma:                 "no-cache",
		Method:                 "GET",
		ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: "1",
		CacheControl:           "max-age=0",
		Connection:             "keep-alive",
		Host:                   domain}

	header.SetUserAgent(`Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.110 Safari/537.36`)

	cr_request, er := client.NewRequest(header.Method, url, nil)

	if er != nil {
		panic(er.Error())
	}
	cr_request.SetHeader(header)

	return client.Do(cr_request.Request, nil)
}

func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

func (client *Client) InitCustomRequest(url string, cHeader map[string]string) (*http.Response, error) {

	cHeader["User-Agent"] = UserAgents[rand.Intn(len(UserAgents))]
	if cHeader["Method"] == "" {
		cHeader["Method"] = "GET"
	}
	cr_request, er := client.NewRequest(cHeader["Method"], url, nil)

	if er != nil {
		return &http.Response{}, er
	}
	cr_request.SetCustomeHeader(cHeader)

	return client.Do(cr_request.Request, nil)
}

// header custome format is map[string]string
func (crRequest *CrRequest) SetCustomeHeader(header map[string]string) {
	for key, value := range header {
		if strings.TrimSpace(key) != "" {
			crRequest.Request.Header.Add(key, value)
		}
	}
}

// setting client with cookie jar
func NewClient() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := http.Client{Timeout: HTTPTIMEOUT, Transport: nil, CheckRedirect: nil, Jar: jar}

	c := &Client{client: &httpClient}
	c.SetProxy()

	return c
}
