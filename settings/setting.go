package settings

/*
Implement go-query to crawl proxy content
*/

import (
	"crypto/tls"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"

	"golang.org/x/net/publicsuffix"
)

// Proxy represent proxy component
type Proxy struct {
	ProxyIP string `json:"proxy_ip"`
	Port    string `json:"port"`
	Schema  string `json:"schema"`
}

// Proxies list of proxy read from file
type Proxies struct {
	ActiveProxy []Proxy `json:"active_proxies"`
}

// Response response instance http
type Response struct {
	*http.Response
}

const (
	ErrProxyPrefix          = "Error, failed on connect proxy, "
	DefaultAccept           = "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3"
	DefaultAcceptLanguage   = "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6"
	DefaultAcceptEncoding   = "gzip,deflate,sdch"
	DefaultInsecCureReqeust = "1"
	DefaultCacheControl     = "max-age=0"
	DefaultConnection       = "keep-alive"
	DefaultSecFetchDest     = "document"
	DefaultSecFetchMode     = "navigate"
	DefaultSecFetchSite     = "none"
	DefaultSecFetchUser     = "?1"
)

// CrRequest Custom new request type for setting header on http.Request
type CrRequest struct {
	*http.Request
}

// ToString return proxy as string
func (proxy *Proxy) ToString() string {
	return proxy.Schema + "://" + proxy.ProxyIP + ":" + proxy.Port
}

// SetHeader set header request from Header
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

// Header header instance on request
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
	SecFetchDest           string
	SecFetchMode           string
	SecFetchSite           string
	SecFetchUser           string
}

// SetUserAgent set user_agent on header
func (header *Header) SetUserAgent(ua string) {
	header.UserAgent = ua
}

// Client represent client http
type Client struct {
	client *http.Client
}

// NewRequest make new request return request type client
func (client *Client) newRequest(method, urlStr string, body io.Reader) (*CrRequest, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, u.String(), body)
	if err != nil {
		return nil, err
	}
	// defer req.Body.Close()

	request := &CrRequest{Request: req}

	return request, nil
}

// SetProxy setting proxy from list proxy read from file on http, https or socks5 proxy
func (client *Client) SetProxy() Proxy {

	condition := true
	var usedProxy Proxy
	for condition {
		transport := &http.Transport{
			DisableCompression:  true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
		var proxyURL *url.URL
		proxyIndex := rand.Intn(len(ProxyList.ActiveProxy))
		usedProxy = ProxyList.ActiveProxy[proxyIndex]
		proxyURL, err := url.Parse(usedProxy.Schema + "://" + usedProxy.ProxyIP + ":" + usedProxy.Port)
		if err != nil {
			condition = true
		} else {
			transport.Proxy = http.ProxyURL(proxyURL)

			if usedProxy.Schema == "socks5" {
				addr := usedProxy.ProxyIP + ":" + usedProxy.Port
				dialer, err := proxy.SOCKS5(
					"tcp",
					addr,
					nil,
					&net.Dialer{
						Timeout:   1 * time.Minute,
						KeepAlive: 30 * time.Second})
				if err != nil {
					condition = true
				} else {
					transport.Dial = dialer.Dial
				}
			}
		}

		client.client = &http.Client{
			Transport: transport}
		return usedProxy
	}
	return usedProxy
}

// Do a request with re-try connect proxy with 10 time
func (client *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// retrieve another proxy 3 time request when failed
	flag := 0
	var proxy Proxy
	erRe := errors.New("")
	for flag < 3 {
		resp, err := client.client.Do(req)
		if err == nil {
			return resp, err

		}
		flag++
		erRe = err
		proxy = client.SetProxy()
	}

	msg := ErrProxyPrefix + proxy.ToString()
	log.Println(msg, erRe.Error())
	return nil, errors.New(msg)
}

// InitRequest init request on fix referer
func (client *Client) InitRequest(url string) (*http.Response, error) {
	header := Header{
		Referrer:       "https://www.google.com.vn/",
		Accept:         "ext/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		AcceptLanguage: "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		Pragma:         "no-cache",
		// Method:         "GET",
		AcceptEncoding: "gzip, deflate, br",
		// ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: "1",
		CacheControl:           "no-cache",
		Connection:             "keep-alive",
		Host:                   "www.similarweb.com",
		SecFetchDest:           DefaultSecFetchDest,
		SecFetchMode:           DefaultSecFetchMode,
		SecFetchSite:           DefaultSecFetchSite,
		SecFetchUser:           DefaultSecFetchUser}

	header.SetUserAgent(UserAgents[rand.Intn(len(UserAgents))])

	csRequest, er := client.newRequest(header.Method, url, nil)

	if er != nil {
		panic(er.Error())
	}
	csRequest.SetHeader(header)

	return client.Do(csRequest.Request, nil)
}

// InitRequest2 init request on custom referer
func (client *Client) InitRequest2(url string, hostname string, domain string) (*http.Response, error) {
	header := Header{
		Referrer:       hostname,
		Accept:         DefaultAccept,
		AcceptLanguage: DefaultAcceptLanguage,
		AcceptEncoding: DefaultAcceptEncoding,
		// Pragma:                 "no-cache",
		// Method:                 "GET",
		// ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: DefaultInsecCureReqeust,
		CacheControl:           DefaultCacheControl,
		Connection:             DefaultConnection,
		Host:                   domain}

	header.SetUserAgent(UserAgents[rand.Intn(len(UserAgents))])

	csRequest, er := client.newRequest(header.Method, url, nil)

	if er != nil {
		panic(er.Error())
	}
	csRequest.SetHeader(header)

	return client.Do(csRequest.Request, nil)
}

func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// InitCustomRequest init new request for client on custome header
func (client *Client) InitCustomRequest(url string, cHeader map[string]string) (*http.Response, error) {

	cHeader["User-Agent"] = UserAgents[rand.Intn(len(UserAgents))]
	if cHeader["Method"] == "" {
		cHeader["Method"] = "GET"
	}
	csRequest, er := client.newRequest(cHeader["Method"], url, nil)

	if er != nil {
		return &http.Response{}, er
	}
	csRequest.SetCustomeHeader(cHeader)

	return client.Do(csRequest.Request, nil)
}

// SetCustomeHeader set header custome format is map[string]string
func (cr *CrRequest) SetCustomeHeader(header map[string]string) {
	for key, value := range header {
		if strings.TrimSpace(key) != "" {
			cr.Request.Header.Add(key, value)
		}
	}
}

// NewClient setting client with cookie jar
func NewClient() *Client {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	httpClient := http.Client{Timeout: HTTPTIMEOUT, Transport: nil, CheckRedirect: nil, Jar: jar}

	c := &Client{client: &httpClient}
	c.SetProxy()

	return c
}
