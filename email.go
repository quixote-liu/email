package email

import (
	"net/smtp"
	"strings"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Addr    string
	body    []byte

	smtpAuth smtp.Auth
}

func (e *Email) auth() smtp.Auth {
	if e.smtpAuth != nil {
		return e.smtpAuth
	}
	addrParts := strings.Split(e.Addr, ":")
	auth := smtp.PlainAuth("", e.From, "osgdkjfzrbgkjbjg", addrParts[0])
	e.smtpAuth = auth
	return e.smtpAuth
}

func (e *Email) render(message string) []byte {
	// contentType := "Content-Type: text/plain;charset=UTF-8"
	// msg := fmt.Sprintf("To:%s\r\nFrom:%s<%s>+\r\nSubject:%s\r\n%s\r\n\r\n%s", e.To[0])
	return []byte("hello, world")
}

func (e *Email) Send(message string) error {
	return smtp.SendMail(e.Addr, e.auth(), e.From, e.To, []byte(message))
}
