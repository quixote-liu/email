package email

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Email struct {
	From     string
	Sender   string
	ReplyTo  []string
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Comments string

	// with host and port.
	Addr string

	text        *bytes.Buffer
	html        *bytes.Buffer
	attachments []*attachment
	auth        smtp.Auth
}

func (e *Email) SetAuth(a smtp.Auth) *Email {
	e.auth = a
	return e
}

func (e *Email) WriteText(text []byte) *Email {
	e.text = bytes.NewBuffer(text)
	return e
}

func (e *Email) WriteHTML(html []byte) *Email {
	e.html = bytes.NewBuffer(html)
	return e
}

func (e *Email) Reset() {
	if e.html != nil {
		e.html.Reset()
	}
	if e.text != nil {
		e.text.Reset()
	}
	e.attachments = make([]*attachment, 0)
}

func (e *Email) AttachFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	ct := mime.TypeByExtension(filepath.Ext(f.Name()))
	basename := filepath.Base(filename)
	return e.attach(f, basename, ct)
}

func (e *Email) attach(r io.Reader, filename, contentType string) error {
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, r); err != nil {
		return err
	}
	a := &attachment{
		filename:    filename,
		contentType: contentType,
		content:     buf,
	}
	a.setDefaultHeader()
	e.attachments = append(e.attachments, a)
	return nil
}

func (e *Email) messageHeaders() (textproto.MIMEHeader, error) {
	header := make(textproto.MIMEHeader, 0)
	header.Set("MIME-Version", "1.0")
	header.Set("From", e.From)
	header.Set("To", strings.Join(e.To, ", "))
	header.Set("Subject", e.Subject)
	header.Set("Date", time.Now().Format(time.RFC3339))
	if e.Sender != "" {
		header.Set("Sender", e.Sender)
	}
	if len(e.ReplyTo) > 0 {
		header.Set("Reply-To", strings.Join(e.ReplyTo, ", "))
	}
	if len(e.CC) > 0 {
		header.Set("CC", strings.Join(e.CC, ", "))
	}
	if len(e.BCC) > 0 {
		header.Set("BCC", strings.Join(e.BCC, ", "))
	}
	if e.Comments != "" {
		header.Set("Comments", e.Comments)
	}

	// set message-id
	mid, err := generateMessageID()
	if err != nil {
		return nil, fmt.Errorf("generate message id failed: %v", err)
	}
	header.Set("Message-Id", mid)

	return header, nil
}

func (e *Email) messageBody() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	defer buf.Reset()
	writer := multipart.NewWriter(buf)
	defer writer.Close()

	headers, err := e.messageHeaders()
	if err != nil {
		return nil, err
	}

	var isMixed = len(e.attachments) > 0
	switch {
	case isMixed:
		headers.Set("Content-Type", "multipart/mixed;\r\n boundary="+writer.Boundary())
	case e.html != nil:
		headers.Set("Content-Type", "text/html; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	default:
		headers.Set("Content-Type", "text/plain; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	}

	if err := e.writeHeadersToBody(buf, headers); err != nil {
		return nil, err
	}

	if e.text != nil {
		if isMixed {
			h := textproto.MIMEHeader{
				"Content-Type":              {"text/plain; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			}
			if _, err := writer.CreatePart(h); err != nil {
				return nil, err
			}
		}
		if _, err := buf.Write(e.text.Bytes()); err != nil {
			return nil, err
		}
	}

	if e.html != nil {
		if isMixed {
			h := textproto.MIMEHeader{
				"Content-Type":              {"text/html; charset=UTF-8"},
				"Content-Transfer-Encoding": {"quoted-printable"},
			}
			if _, err := writer.CreatePart(h); err != nil {
				return nil, err
			}
		}
		if _, err := buf.Write(e.html.Bytes()); err != nil {
			return nil, err
		}
	}

	_, _ = buf.WriteString("\r\n")

	for _, at := range e.attachments {
		if _, err := writer.CreatePart(at.header); err != nil {
			return nil, err
		}
		at.writeBase64To(buf)
	}

	return buf.Bytes(), nil
}

func (e *Email) writeHeadersToBody(buf *bytes.Buffer, headers textproto.MIMEHeader) error {
	for field, values := range headers {
		for _, subval := range values {
			_, _ = buf.WriteString(field)
			buf.WriteString(": ")
			switch {
			case field == "Content-Type" || field == "Content-Dispostion":
				_, _ = buf.WriteString(subval)
			case field == "From" || field == "To" || field == "CC" || field == "BCC":
				parts := strings.Split(subval, ",")
				for i, v := range parts {
					addr, err := mail.ParseAddress(v)
					if err != nil {
						return err
					}
					parts[i] = addr.Address
				}
				_, _ = buf.WriteString(strings.Join(parts, ", "))
			default:
				_, _ = buf.WriteString(mime.QEncoding.Encode("UTF-8", subval))
			}
			_, _ = buf.WriteString("\r\n")
		}
	}
	_, _ = buf.WriteString("\r\n")
	return nil
}

func (e *Email) Send() error {
	to := make([]string, 0)
	to = append(append(append(to, e.To...), e.CC...), e.BCC...)
	for i, t := range to {
		addr, err := mail.ParseAddress(t)
		if err != nil {
			return err
		}
		to[i] = addr.Address
	}

	sender, err := e.parseSender()
	if err != nil {
		return err
	}

	bytes, err := e.messageBody()
	if err != nil {
		return err
	}

	return smtp.SendMail(e.Addr, e.auth, sender, to, bytes)
}

func (e *Email) parseSender() (string, error) {
	if e.Sender != "" {
		addr, err := mail.ParseAddress(e.Sender)
		if err != nil {
			return "", err
		}
		return addr.Address, nil
	}

	addr, err := mail.ParseAddress(e.From)
	if err != nil {
		return "", err
	}
	return addr.Address, nil
}
