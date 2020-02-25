package settings

/*
Implement selenium to crawler proxy and user-agents
*/

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"../utils"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

var UserAgents = SGetUserAgents()
var ProxyList = SGetActiveProxy()

// var CrawlerProfile = "setting_profile_driver"
//var CacheDir = "setting_cache_dir"

// get user-agent list from file
func SGetUserAgents() []string {
	var user_agents []string

	//File not exist in path, implement crawl proxy
	if _, err := os.Stat(USERAGENTFILE); os.IsNotExist(err) {
		user_agents, er := SCrawlerUserAgents("Chrome")
		if er != nil {
			panic(er.Error())
		} else {
			return user_agents
		}
	} else {
		ua_file, err := os.Open(USERAGENTFILE)
		if err != nil {
			panic(err.Error())
		}
		defer ua_file.Close()

		err = json.NewDecoder(ua_file).Decode(&user_agents)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(len(user_agents))
		if len(user_agents) == 0 {
			fmt.Println("enter crawler")
			user_agents, err = SCrawlerUserAgents("Chrome")
			if err != nil {
				panic(err.Error())
			}
		}
		return user_agents
	}
}

// run crawler user-agent and write to file
func SCrawlerUserAgents(browser string) ([]string, error) {
	caps := selenium.Capabilities(map[string]interface{}{"browserName": "chrome"})

	// connect to selenium Standalone alone (run on java jar package)
	webDriver, err := InitNewRemote(caps, utils.STANDALONESERVER)
	if err != nil {
		fmt.Printf("Failed to open session: %s\n", err)
		return []string{}, err
	}
	defer webDriver.Quit()

	err = webDriver.Get(UserAgentString[0] + browser)
	var uas []string
	if err != nil {
		return []string{}, nil
	} else {
		listAgents, err := webDriver.FindElements(selenium.ByCSSSelector, "#liste ul li a")
		if err != nil {
			return []string{}, err
		} else {
			for index, agent := range listAgents {
				a_agent, err := agent.Text()
				if err == nil {
					uas = append(uas, a_agent)
				}
				if index > 50 {
					break
				}
			}
		}
	}
	fmt.Println("User agentssssssssssssssssssssssssss", len(uas))
	result, err := json.Marshal(uas)
	if err != nil {
		panic(err.Error())
	}

	ua_file, err := os.OpenFile(USERAGENTFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err.Error())
	}
	defer ua_file.Close()
	ua_file.Write(result)

	return uas, err

}

// get proxy list from proxy file
func SGetActiveProxy() Proxies {

	var active_proxy Proxies

	//File not exist in path, implement crawl proxy
	if _, err := os.Stat(PROXYFILE); os.IsNotExist(err) {
		active_proxy, er := SCrawlerProxy()
		if er != nil {
			panic(er.Error())
		} else {
			return active_proxy
		}
	} else {
		proxy_file, err := os.Open(PROXYFILE)
		if err != nil {
			panic(err.Error())
		}
		defer proxy_file.Close()

		err = json.NewDecoder(proxy_file).Decode(&active_proxy)
		if err != nil {
			panic(err.Error())
		}
		fmt.Println(len(active_proxy.ActiveProxy))
		if len(active_proxy.ActiveProxy) == 0 {
			fmt.Println("enter crawler")
			active_proxy, err = SCrawlerProxy()
			if err != nil {
				panic(err.Error())
			}
		}
		return active_proxy
	}
}

// run crawler proxy and write to file
func SCrawlerProxy() (Proxies, error) {

	caps := selenium.Capabilities(map[string]interface{}{"browserName": "chrome"})

	// connect to selenium Standalone alone (run on java jar package)
	webDriver, err := InitNewRemote(caps, utils.STANDALONESERVER)
	if err != nil {
		fmt.Printf("Failed to open session: %s\n", err)
		return Proxies{}, err
	}
	defer webDriver.Quit()

	err = webDriver.Get(FreeProxy[5])
	if err != nil {
		return Proxies{}, err
	}

	tr_elements, err := webDriver.FindElements(selenium.ByCSSSelector, "#tbl_proxy_list tbody tr")
	if err != nil {
		return Proxies{}, err
	}
	proxies := Proxies{}
	for _, tr_element := range tr_elements {
		fmt.Println("Find elementssssssssssssssssssssssss")

		// only get content from data proxy id attribute
		_, att_er := tr_element.GetAttribute("data-proxy-id")
		if att_er == nil {
			fmt.Println("Find attrrrrrrrrrrrrrrrrrrr")
			var proxy_row []string
			tds, err := tr_element.FindElements(selenium.ByCSSSelector, "td")
			if err == nil {
				for _, td := range tds {
					piece, err := td.Text()
					if err == nil {
						proxy_row = append(proxy_row, piece)
					}
				}
			}
			fmt.Println("proxy rowwwwwwwwwwwwwwwwwwwww", proxy_row)

			if len(proxy_row) > 0 {
				proxy_ip := strings.TrimSpace(proxy_row[0])
				port := strings.TrimSpace(proxy_row[1])
				schema := "http"
				fmt.Println("proxyyyyyyyyyyyyyyyy", proxy_ip)

				proxy := Proxy{ProxyIP: proxy_ip, Port: port, Schema: schema}
				proxies.ActiveProxy = append(proxies.ActiveProxy, proxy)
			}
		}
	}
	fmt.Println("proxiessssssssssssssssssssssssss", proxies)
	result, err := json.Marshal(proxies)
	if err != nil {
		panic(err.Error())
	}

	proxy_file, err := os.OpenFile(PROXYFILE, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		panic(err.Error())
	}
	defer proxy_file.Close()
	proxy_file.Write(result)

	return proxies, err

}

// selenium use proxy
func SetChomeCapabilities() selenium.Capabilities {

	var agrs []string
	agrs = append(agrs, fmt.Sprintf("--user-agent=%s", UserAgents[rand.Intn(len(UserAgents))]))
	// have no use UI chrome windows
	agrs = append(agrs, "--headless")
	// agrs = append(agrs, "--start-maximized")
	// agrs = append(agrs, "--disk-cache-size=500000")

	// using profile craw with no loading images
	// agrs = append(agrs, fmt.Sprintf("--profile-directory=%s", CrawlerProfile))

	// agrs = append(agrs, fmt.Sprintf("--disk-cache-dir=%s", CacheDir))
	chrome_caps := chrome.Capabilities{Args: agrs}
	caps := selenium.Capabilities(map[string]interface{}{"browserName": "chrome"})
	caps.AddChrome(chrome_caps)

	setting_proxy := ProxyList.ActiveProxy[rand.Intn(len(ProxyList.ActiveProxy))]
	proxy := selenium.Proxy{AutoconfigURL: setting_proxy.Schema + ":\\" + setting_proxy.ProxyIP + ":" + setting_proxy.Port, Type: "manual"}

	caps.AddProxy(proxy)
	return caps

}

// selenium not use proxy
func SetFreeProxyChomeCapabilities() selenium.Capabilities {

	var agrs []string
	agrs = append(agrs, fmt.Sprintf("--user-agent=%s", UserAgents[rand.Intn(len(UserAgents))]))

	// have no use UI chrome windows
	agrs = append(agrs, "--headless")
	agrs = append(agrs, "--start-maximized")
	agrs = append(agrs, "--no-sandbox")
	agrs = append(agrs, "--disk-cache-size=500000")

	// using profile craw with no loading images
	// agrs = append(agrs, fmt.Sprintf("--profile-directory=%s", CrawlerProfile))
	// agrs = append(agrs, fmt.Sprintf("--disk-cache-dir=%s", CacheDir))
	chrome_caps := chrome.Capabilities{Args: agrs}
	caps := selenium.Capabilities(map[string]interface{}{"browserName": "chrome"})
	caps.AddChrome(chrome_caps)

	return caps

}

// init web driver with setting caps for crawler
func InitNewRemote(caps selenium.Capabilities, url string) (selenium.WebDriver, error) {
	webDriver, err := selenium.NewRemote(caps, url)
	if err != nil {
		return nil, err
	}
	return webDriver, nil
}
