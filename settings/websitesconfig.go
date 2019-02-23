package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

type WebsiteConfig struct {
	Domain         string `json:"domain"`
	Hostname       string `json:"host_name"`
	URL            string `json:"url"`
	ContentStruct  string `json:"content_class"`
	CategoryStruct string `json:"category_class"`
	DateStruct     string `json:"date_class"`
	SpecialHeader  bool   `json:"special_header"`
}

type Config struct {
	Config []WebsiteConfig `json:"config"`
}

func readConfig() ([]WebsiteConfig, error) {
	var config Config
	var websites []WebsiteConfig
	//File not exist in path
	if _, err := os.Stat(CONFIGFILE); os.IsNotExist(err) {
		return websites, err
	}
	config_file, err := os.Open(CONFIGFILE)
	if err != nil {
		panic(err.Error())
	}
	defer config_file.Close()

	err = json.NewDecoder(config_file).Decode(&config)
	if err != nil {
		panic(err.Error())
	}
	websites = config.Config
	if len(websites) == 0 {
		fmt.Println("Info: Your config is empty, please add more config")
		panic(errors.New("Your config is empty, please add more config"))
	}
	return websites, err
}

func SetConfig() map[string]WebsiteConfig {
	websites, er := readConfig()
	fmt.Println(websites)
	if er != nil {
		panic(er.Error())
	}

	var config = map[string]WebsiteConfig{}

	for _, website := range websites {
		config[website.Domain] = website
	}
	return config
}
