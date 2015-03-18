package main

import (
	"html"
	"html/template"
	"strings"
	"time"
)

func para(text string) template.HTML {
	text = html.EscapeString(text)
	text = strings.Replace(text, "\r", "", -1)
	text = strings.Replace(text, "\n\n", "</p><p>", -1)
	text = strings.Replace(text, "\n", "<br>", -1)
	return template.HTML(text)
}

func formatTime(t time.Time) string {
	return t.In(auckland).Format("2-Feb-2006 3:04pm")
}

func obfEmail(email, label string) template.HTML {
	// Assumes strings are ascii
	emailBytes := []byte(email)
	for i := 0; i < len(emailBytes); i++ {
		emailBytes[i] = emailBytes[i] - 1
	}

	obfEmail := string(emailBytes)
	return template.HTML("<a class=\"obf-email\" href=\"#\" data-obf-email=\"" +
		html.EscapeString(obfEmail) + "\" data-obf-email-label=\"" +
		html.EscapeString(label) + "\">&nbsp;</a>")
}

func siteURL() string {
	return Config.SiteURL
}
