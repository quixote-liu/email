package email

import (
	"bytes"
	"errors"
	"io"
	"mime"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
)

type Email struct {
	ReplyTo     []string
	From        string
	To          []string
	Bcc         []string
	Cc          []string
	Subject     string
	Text        []byte // Plaintext message (optional)
	HTML        []byte // Html message (optional)
	Sender      string // override From as SMTP envelope sender (optional)
	Headers     textproto.MIMEHeader
	Attachments []*attachment
	ReadReceipt []string
}

func NewEmail() *Email {
	return &Email{Headers: textproto.MIMEHeader{}}
}

func (e *Email) Attach(r io.Reader, filename, contentType string) error {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}
	a := &attachment{
		filename:    filename,
		contentType: contentType,
		header:      textproto.MIMEHeader{},
		content:     buf.Bytes(),
	}
	e.Attachments = append(e.Attachments, a)
	return nil
}

func (e *Email) AttachFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	ct := mime.TypeByExtension(filepath.Ext(f.Name()))
	basename := filepath.Base(filename)
	return e.Attach(f, basename, ct)
}

func (e *Email) Send(addr string, a smtp.Auth) error {
	to := make([]string, 0, len(e.To)+len(e.Cc)+len(e.Bcc))
	to = append(append(append(to, e.To...), e.Cc...), e.Bcc...)
	for i, t := range to {
		addr, err := mail.ParseAddress(t)
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}
	if e.From == "" || len(to) == 0 {
		return errors.New("sender or recevier is missing")
	}
	sender, err := e.parseSender()
	if err != nil {
		return err
	}
	msg, err := e.bytes()
	if err != nil {
		return err
	}
	return smtp.SendMail(addr, a, sender, to, msg)
}

func (e *Email) parseSender() (string, error) {
	if e.Sender != "" {
		addr, err := mail.ParseAddress(e.Sender)
		if err != nil {
			return "", err
		}
		return addr.Address, nil
	} else {
		addr, err := mail.ParseAddress(e.From)
		if err != nil {
			return "", err
		}
		return addr.Address, nil
	}
}

func (e *Email) bytes() ([]byte, error) {
	// TODO: optimize.
	return nil, nil
}
