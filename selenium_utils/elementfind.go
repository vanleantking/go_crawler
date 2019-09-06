package selenium_utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

// @TODO: implement findElements list of element by waiting

const (
	NoSuchElement        = "no such element"
	StaleElementRef      = "stale element reference"
	ElementNotVisible    = "element not visible"
	Unknown              = "unknown error"
	ElementNotSelectable = "element is not selectable"
	TimeOut              = "timeout"
	InvalidSelector      = "invalid selector"
	Others               = "others"
)

var ErrorServRep = map[string]int{
	"invalid session ID":          1,
	"no such element":             2,
	"no such frame":               3,
	"unknown command":             4,
	"stale element reference":     5,
	"element not visible":         6,
	"invalid element state":       7,
	"unknown error":               8,
	"element is not selectable":   9,
	"javascript error":            10,
	"xpath lookup error":          11,
	"timeout":                     12,
	"no such window":              13,
	"invalid cookie domain":       14,
	"unable to set cookie":        15,
	"unexpected alert open":       16,
	"no alert open":               17,
	"script timeout":              18,
	"invalid element coordinates": 19,
	"invalid selector":            20}

type ElementFindUtils struct {
	Driver  selenium.WebDriver
	timeout int64
	Wait    int
	Retry   int
	curTime int64
}

// NewEFU return ElementFindUtils instance
func NewEFU(driver selenium.WebDriver, wait int) *ElementFindUtils {
	efu := &ElementFindUtils{
		Driver:  driver,
		Wait:    wait,
		Retry:   3,
		curTime: time.Now().Unix()}

	efu.setTimeOut(wait)
	return efu
}

func (efu *ElementFindUtils) setTimeOut(timeout int) {
	efu.timeout = efu.curTime + int64(timeout)
}

// WaitElementWTimeOut return element after timeout
func (efu *ElementFindUtils) WaitElementWTimeOut(parentElement selenium.WebElement,
	cssSelector string, timeout int64) (selenium.WebElement, error) {
	var elm selenium.WebElement
	var er error

	for true {
		elm, er = efu.findElement(parentElement, cssSelector)
		if er == nil {
			return elm, nil
		}
		er = checkErrorType(er)
		if er.Error() == Others || er.Error() == Unknown || er.Error() == NoSuchElement {
			return nil, er
		}
		time.Sleep(time.Duration(timeout) * time.Second)
		if time.Now().Unix() > efu.timeout {
			er = errors.New(TimeOut)
			break
		}
	}
	return nil, er
}

// WaitUntilClickable
func (efu *ElementFindUtils) WaitUntilClickable(element selenium.WebElement,
	cssSelector string, index int) (bool, error) {
	var er error
	for true {
		// get element
		// check element is visible
		er := element.Click()
		if er == nil {
			return true, nil
		}

		er = checkErrorType(er)
		if er != nil && er.Error() == Others {
			return false, er
		}

		// wait until some times
		time.Sleep(time.Duration(efu.Wait) * time.Second)

		// execute script after wait
		elmIndexExe := ""
		if index < 0 {
			elmIndexExe = "$('" + cssSelector + "')" + ".click()"
		} else {
			elmIndexExe = "$('" + cssSelector + "')" + "[" + strconv.Itoa(index) + "]" + ".click()"
		}

		_, er = efu.Driver.ExecuteScript(elmIndexExe, []interface{}{})
		if er == nil {
			return true, nil
		}
		// return error if error is [unknown, others]
		er = checkErrorType(er)
		if er != nil && (er.Error() == Others || er.Error() == NoSuchElement) {
			return false, er
		}
		if time.Now().Unix() > efu.timeout {
			er = errors.New("timeout")
			return false, er
		}

		time.Sleep(time.Duration(efu.Wait) * time.Second)
	}
	return true, er
}

// GetElementRetrieve return element after retry
func (efu *ElementFindUtils) GetElementRetrieve(parentElement selenium.WebElement,
	cssSelector string) (selenium.WebElement, error) {

	var elm selenium.WebElement
	var er error

	for retrieved := 0; retrieved < efu.Retry; retrieved++ {
		elm, er = efu.findElement(parentElement, cssSelector)
		if er == nil {
			return elm, nil
		}
		time.Sleep(time.Duration(efu.Wait) * time.Second)
	}
	return elm, er
}

// findElement return element by css selector
func (efu *ElementFindUtils) findElement(parentElement selenium.WebElement,
	cssSelector string) (selenium.WebElement, error) {
	elm, er := parentElement.FindElement(selenium.ByCSSSelector, cssSelector)
	return elm, er
}

func checkErrorType(er error) error {
	errString := er.Error()
	for errStr, erType := range ErrorServRep {
		if strings.Contains(errString, errStr) {
			switch erType {
			case 2:
				return errors.New(NoSuchElement)
			case 5:
				return errors.New(StaleElementRef)
			case 8:
				return errors.New(Unknown)
			case 6:
				return errors.New(ElementNotVisible)
			case 9:
				return errors.New(ElementNotSelectable)
			case 11:
				return errors.New(TimeOut)
			case 20:
				return errors.New(InvalidSelector)
			default:
				return errors.New(Others)
			}
		}
	}
	return errors.New(Unknown)
}
