package utils

import (
	"github.com/tebeka/selenium"
)

// selenium Get content news from list of class split by "|" charater from DB
func SeleniumGetContentFromClass(class string, doc selenium.WebDriver) (string, error) {
	var elment selenium.WebElement
	elment, err := doc.FindElement(selenium.ByCSSSelector, class)
	content := ""
	if err == nil {
		content, err = elment.Text()
		return content, err
	}
	return content, err
}
