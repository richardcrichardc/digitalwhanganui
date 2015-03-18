package main

import (
	"encoding/json"
	"io/ioutil"
)

var Config struct {
	SiteURL                                                string
	Debug                                                  bool
	TemplateDir, EmailTemplateDir, PublicDir               string
	SMTPServer, SMTPUser, SMTPPassword                     string
	AdminEmailAddress, ErrorEmailAddress, FromEmailAddress string
}

func loadConfig() {
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(configData, &Config)
	if err != nil {
		panic(err)
	}
}
