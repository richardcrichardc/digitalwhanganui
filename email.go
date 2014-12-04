package main

import (
	"bytes"
	"fmt"
	"gopkg.in/alexcesaro/quotedprintable.v1"
	"net/smtp"
	"path/filepath"
	"text/template"
)

var shortFromEmail = "digitalwhanganui@digitalwhanganui.org.nz"
var fromEmail = "DigitalWhanganui <" + shortFromEmail + ">"
var server = "bakerloo.richardc.net:587"
var auth = smtp.CRAMMD5Auth("digitalwhanganui@digitalwhanganui.org.nz", "apodacaGritz")

var emailTemplates *template.Template

func sendMail(to, subject, template string, data map[string]string) {
	checkHeader(to)
	checkHeader(subject)

	var buf bytes.Buffer
	buf.WriteString("From: " + fromEmail + "\r\n")
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

	fmt.Println(to, subject, template, data)
	fmt.Println(string(msg))

	// TODO Fix email args - almost certainly wrong but works
	err = smtp.SendMail(server, auth, fromEmail, []string{to}, msg)
	if err != nil {
		panic(err)
	}

}

func (l *Listing) FullAdminEmail() string {
	return l.AdminFirstName + " " + l.AdminLastName + " <" + l.AdminEmail + ">"
}

func checkHeader(header string) {
	// TODO
}

func init() {
	var dir string
	if fileExists("email-templates") {
		dir = "email-templates"
	} else {
		dir = "/usr/local/share/digitalwhanganui/email-templates"
	}

	pattern := filepath.Join(dir, "*.tmpl")
	emailTemplates = template.Must(template.ParseGlob(pattern))

}
