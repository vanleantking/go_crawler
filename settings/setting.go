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

const (
	ErrProxyPrefix = "Error, failed on connect proxy, "
)

// Custom new request type for setting header on http.Request
type CrRequest struct {
	*http.Request
}

func (proxy *Proxy) ToString() string {
	return proxy.Schema + "://" + proxy.ProxyIP + ":" + proxy.Port
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

	cr_request := &CrRequest{Request: req}

	return cr_request, nil
}

// SetProxy setting proxy on http, https or socks5 proxy
func (client *Client) SetProxy() Proxy {

	condition := true
	var used_proxy Proxy
	for condition {
		transport := &http.Transport{
			DisableCompression:  true,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		}
		var proxyUrl *url.URL
		proxy_index := rand.Intn(len(ProxyList.ActiveProxy))
		used_proxy = ProxyList.ActiveProxy[proxy_index]
		proxyUrl, err := url.Parse(used_proxy.Schema + "://" + used_proxy.ProxyIP + ":" + used_proxy.Port)
		if err != nil {
			condition = true
		} else {
			transport.Proxy = http.ProxyURL(proxyUrl)

			if used_proxy.Schema == "socks5" {
				addr := used_proxy.ProxyIP + ":" + used_proxy.Port

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
		return used_proxy
	}
	return used_proxy
}

// Do a request with re-try connect proxy with 10 time
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	// retrieve another proxy 3 time request when failed
	flag := 0
	var proxy Proxy
	re_error := errors.New("")
	for flag < 10 {
		resp, err := c.client.Do(req)
		if err == nil {
			return resp, err

		}
		flag++
		re_error = err
		proxy = c.SetProxy()
	}

	msg := ErrProxyPrefix + proxy.ToString()
	log.Println(msg, re_error.Error())
	return nil, errors.New(msg)
}

// InitRequest init request on fix referer
func (client *Client) InitRequest(url string) (*http.Response, error) {
	header := Header{
		Referrer:       "https://www.google.com.vn/",
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
		AcceptLanguage: "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		// Pragma:                 "no-cache",
		Method: "GET",
		// ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: "1",
		CacheControl:           "max-age=0",
		Connection:             "keep-alive"}

	header.SetUserAgent(UserAgents[rand.Intn(len(UserAgents))])

	cr_request, er := client.newRequest(header.Method, url, nil)

	if er != nil {
		panic(er.Error())
	}
	cr_request.SetHeader(header)

	return client.Do(cr_request.Request, nil)
}

// InitRequest2 init request on custom referer
func (client *Client) InitRequest2(url string, host_name string, domain string) (*http.Response, error) {
	header := Header{
		Referrer:       host_name,
		Accept:         "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
		AcceptLanguage: "vi,en-GB;q=0.9,en;q=0.8,en-US;q=0.7,ja;q=0.6",
		// Pragma:                 "no-cache",
		// Method:                 "GET",
		// ContentType:            "text/html; charset=utf-8",
		UpdateInsecCureRequest: "1",
		CacheControl:           "max-age=0",
		Connection:             "keep-alive",
		Host:                   domain}

	header.SetUserAgent(UserAgents[rand.Intn(len(UserAgents))])

	cr_request, er := client.newRequest(header.Method, url, nil)

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

// InitCustomRequest init new request for client on custome header
func (client *Client) InitCustomRequest(url string, cHeader map[string]string) (*http.Response, error) {

	cHeader["User-Agent"] = UserAgents[rand.Intn(len(UserAgents))]
	if cHeader["Method"] == "" {
		cHeader["Method"] = "GET"
	}
	cr_request, er := client.newRequest(cHeader["Method"], url, nil)

	if er != nil {
		return &http.Response{}, er
	}
	cr_request.SetCustomeHeader(cHeader)

	return client.Do(cr_request.Request, nil)
}

// SetCustomeHeader set header custome format is map[string]string
func (crRequest *CrRequest) SetCustomeHeader(header map[string]string) {
	for key, value := range header {
		if strings.TrimSpace(key) != "" {
			crRequest.Request.Header.Add(key, value)
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
