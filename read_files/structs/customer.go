package structs

import (
	"errors"
	"regexp"
	"strings"
)

const (
	Email string = "^(((([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|((\\x22)((((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(([\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(\\([\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(((\\x20|\\x09)*(\\x0d\\x0a))?(\\x20|\\x09)+)?(\\x22)))@((([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|\\.|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])([a-zA-Z]|\\d|-|_|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*([a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
)

var VNMobilePrefix = map[string]string{
	// ViettelPrefix
	"0162": "032",
	"0163": "033",
	"0164": "034",
	"0165": "035",
	"0166": "036",
	"0167": "037",
	"0168": "038",
	"0169": "039",
	//MobilePrefix
	"0120": "070",
	"0121": "079",
	"0122": "077",
	"0126": "076",
	"0128": "078",
	//VinaPrefix
	"0123": "083",
	"0124": "084",
	"0125": "085",
	"0127": "081",
	"0129": "082",
	//VinaMobilePrefix
	"0186": "056",
	"0188": "058",
	//GMobilePrefix
	"0199": "059"}

type Customer struct {
	CustomerName string `json:"customer_name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Address      string `json:"address"`
	Note         string `json:"note"`
	CookieName   string `json:"cookie_name"`
}

func (customer *Customer) SetCustomerImport(record []string, header map[string]int) error {
	var err error
	if _, ok := header["address"]; ok {
		customer.Address = strings.TrimSpace(record[header["address"]])
	}

	if _, ok := header["email"]; ok {
		customer.Email, err = getEmailCustomer(strings.TrimSpace(record[header["email"]]))
		if err != nil {
			return err
		}
	}

	if _, ok := header["phone"]; ok {
		customer.Phone = getRealPhoneNumber(strings.TrimSpace(record[header["phone"]]))
	}

	if _, ok := header["name"]; ok {
		customer.CustomerName = strings.TrimSpace(record[header["name"]])
	}

	if _, ok := header["note"]; ok {
		customer.Note = strings.TrimSpace(record[header["note"]])
	}
	is_empty := strings.TrimSpace(customer.Email + customer.Phone)
	if is_empty == "" {
		return errors.New("Email or Phone is required")
	}
	return nil
}

// process phone number
func getRealPhoneNumber(number string) string {
	//replace any space in number phone
	number = strings.Replace(number, " ", "", -1)

	//check exist +84 in number, replace +84 with 0
	rephone := regexp.MustCompile(`\+?84`)
	flag := rephone.FindString(number)
	if flag != "" {
		number = rephone.ReplaceAllLiteralString(number, "")
		number = "0" + number
	}

	// update phone number 11 digits to 10 digits
	if len(number) == 11 {
		prefix := number[:4]
		if _, ok := VNMobilePrefix[prefix]; ok {
			return strings.Replace(number, prefix, VNMobilePrefix[prefix], -1)
		}
	}

	if len(number) == 9 {
		number = "0" + number
	}
	return number
}

func getEmailCustomer(email string) (string, error) {
	if strings.TrimSpace(email) == "" {
		return "", nil
	}

	// this regex fail on empty email
	Re := regexp.MustCompile(Email)
	if Re.MatchString(email) == true {
		return email, nil
	}
	return "", errors.New("Email wrong format")
}
