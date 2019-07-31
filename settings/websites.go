package settings

import "time"

const (
	USERAGENTFILE = "../config/user_agents.json"
	PROXYFILE     = "../config/proxy.json"
	CONFIGFILE    = "../config/config.json"
	HTTPTIMEOUT   = time.Duration(5 * time.Minute)
)

var FreeProxy = []string{"http://spys.one/free-proxy-list/ALL/", // complicated structure, not crawler, port write with document.write <= run script on web
	"http://www.freeproxylists.net/?c=VN&pt=&pr=&a%5B%5D=0&a%5B%5D=1&a%5B%5D=2&u=0", // capcha restriction, not craw yet
	"http://www.gatherproxy.com/proxylist/country/?c=Vietnam",                       // post method, port and ip get with document.write('163.44.206.148') document.write(gp.dep('22B8'))
	"https://www.proxydocker.com/en/proxylist/country/Vietnam",                      // free at first page, register to get more
	"https://premproxy.com/proxy-by-country/Vietnam-01.htm",                         // port write with document.write <= run script on web
	"https://www.proxynova.com/proxy-server-list/country-vn/",
	"https://free-proxy-list.net/",
	"http://www.freeproxylists.net/vn.html"}

var UserAgentString = []string{"http://useragentstring.com/pages/useragentstring.php?name=",
	"https://udger.com/resources/ua-list/browser-detail?browser=Chrome",
	"https://developers.whatismybrowser.com/useragents/explore/software_name/chrome/2"}

// var HttpTimeout2 = time.Duration(5 * time.Second)

// link get content ajax caffebiz
// http://cafebiz.vn/ajax/loadListNews-0-7.chn 7 - page

// genk
// http://genk.vn/ajax-home/page-4/20190527115243573__20190528121147901__20190528112006011__20190528142944058__20190528104047589.chn
