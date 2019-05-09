package utils

import (
	"strings"
)

func GetDomainName(hostname string) string {
	return strings.Replace(hostname, "www.", "", -1)
}
