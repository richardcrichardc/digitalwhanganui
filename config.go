package main

import (
	"encoding/json"
	"io/ioutil"
)

var Config struct {
	SiteURL                                  string
	Debug                                    bool
	TemplateDir, EmailTemplateDir, PublicDir string
	SMTPServer, SMTPUser, SMTPPassword       string
	AdminEmailAddress, FromEmailAddress      string
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

/*
// Decide where to load templates from
var templateDir, publicDir string

if fileExists("templates") {
  templateDir = "templates"
} else {
  templateDir = "/usr/local/share/digitalwhanganui/templates"
}

if fileExists("public") {
  publicDir = "public"
} else {
  publicDir = "/usr/local/share/digitalwhanganui/public"
}
*/
