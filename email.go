package main

import (
	"bytes"
	"log"
	"net/smtp"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/alexcesaro/quotedprintable.v1"
)

var emailTemplates *template.Template

func administratorEmail() string {
	return "DigitalWhanganui <" + Config.AdminEmailAddress + ">"
}

func fromEmail() string {
	return "DigitalWhanganui <" + Config.FromEmailAddress + ">"
}

func sendMail(to, subject, template string, data map[string]string) {
	checkHeaderText(to)
	checkHeaderText(subject)

	var buf bytes.Buffer
	buf.WriteString("From: " + fromEmail() + "\r\n")
	buf.WriteString("To: " + to + "\r\n")
	buf.WriteString("Subject: " + subject + "\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buf.WriteString("\r\n")

	QPWriter := quotedprintable.NewEncoder(&buf)

	err := emailTemplates.ExecuteTemplate(QPWriter, template, data)
	if err != nil {
		panic(err)
	}
	msg := buf.Bytes()

	auth := smtp.CRAMMD5Auth(Config.SMTPUser, Config.SMTPPassword)
	err = smtp.SendMail(Config.SMTPServer, auth, Config.FromEmailAddress, []string{to}, msg)
	if err != nil {
		panic(err)
	}

}

func (l *Listing) FullAdminEmail() string {
	return l.AdminFirstName + " " + l.AdminLastName + " <" + l.AdminEmail + ">"
}

// Ensure headerText is not being used to inject additional email headers
func checkHeaderText(headerText string) {
	if strings.ContainsAny(headerText, "\n\r") {
		panic("Possible Email header injection: " + headerText)
	}
}

func emailErrorMsg(msg string, logger *log.Logger) {
	defer func() {
		if err := recover(); err != nil {
			logger.Println("Panic when emailing error message:", err)
		}
	}()

	sendMail("richardc+digitalwhanganui@richardc.net", "Digital Whanganui Error", "error.tmpl", map[string]string{"error": msg})
}

func initEmail() {
	pattern := filepath.Join(Config.EmailTemplateDir, "*.tmpl")
	emailTemplates = template.Must(template.ParseGlob(pattern))

}
