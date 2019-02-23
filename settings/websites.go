package settings

import "time"

const (
	USERAGENTFILE = "./config/user_agents.json"
	PROXYFILE     = "./config/proxy.json"
	CONFIGFILE    = "./config/config.json"
)

var FreeProxy = []string{"http://spys.one/free-proxy-list/VN/", // complicated structure, not crawler, port write with document.write <= run script on web
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

var HttpTimeout = time.Duration(20 * time.Second)
var HttpTimeout2 = time.Duration(30 * time.Second)
