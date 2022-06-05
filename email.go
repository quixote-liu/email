package email

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	buf := bytes.NewBuffer(make([]byte, 0, 4096))

	headers, err := e.headers()
	if err != nil {
		return nil, err
	}

	htmlAttachments, otherAttachments := e.categorizeAttachments()
	if len(e.HTML) == 0 && len(htmlAttachments) > 0 {
		return nil, fmt.Errorf("there are HTML attachments, but no HTML body")
	}

	var (
		isMixed       = len(otherAttachments) > 0
		isAlternative = len(e.Text) > 0 && len(e.HTML) > 0
		isRelated     = len(e.HTML) > 0 && len(htmlAttachments) > 0
	)

	var w *multipart.Writer
	if isMixed || isAlternative || isRelated {
		w = multipart.NewWriter(buf)
	}
	switch {
	case isMixed:
		headers.Set("Content-Type", "multipart/mixed;\r\n boundary="+w.Boundary())
	case isAlternative:
		headers.Set("Content-Type", "multipart/alternative;\r\n boundary="+w.Boundary())
	case isRelated:
		headers.Set("Content-Type", "multipart/related;\r\n boundary="+w.Boundary())
	case len(e.HTML) > 0:
		headers.Set("Content-Type", "text/html; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	default:
		headers.Set("Content-Type", "text/plain; charset=UTF-8")
		headers.Set("Content-Transfer-Encoding", "quoted-printable")
	}
	e.headersToBytes(buf, headers)
	_, err = io.WriteString(buf, "\r\n")
	if err != nil {
		return nil, err
	}

	if len(e.Text) > 0 || len(e.HTML) > 0 {
		var subWriter *multipart.Writer

		if isMixed && isAlternative {
			header := textproto.MIMEHeader{
				"Content-Type": {"multipart/alternative;\r\n boundary=", subWriter.Boundary()},
			}
			if _, err := w.CreatePart(header); err != nil {
				return nil, err
			}
		} else {
			subWriter = w
		}
		if len(e.Text) > 0 {
			if err := e.writeMessage(buf, e.Text, isMixed || isAlternative, "text/plain", subWriter); err != nil {
				return nil, err
			}
		}
		if len(e.HTML) > 0 {
			messageWriter := subWriter
			var relatedWriter *multipart.Writer
			if (isMixed || isAlternative) && len(htmlAttachments) > 0 {
				relatedWriter = multipart.NewWriter(buf)
				header := textproto.MIMEHeader{
					"Content-Type": {"multipart/related;\r\n boundary=" + relatedWriter.Boundary()},
				}
				if _, err := subWriter.CreatePart(header); err != nil {
					return nil, err
				}

				messageWriter = relatedWriter
			} else if isRelated && len(htmlAttachments) > 0 {
				relatedWriter = w
				messageWriter = w
			}
			// Write the HTML
			if err := e.writeMessage(buf, e.HTML, isMixed || isAlternative || isRelated, "text/html", messageWriter); err != nil {
				return nil, err
			}
			if len(htmlAttachments) > 0 {
				for _, a := range htmlAttachments {
					a.setDefaultHeaders()
					ap, err := relatedWriter.CreatePart(a.header)
					if err != nil {
						return nil, err
					}
					// Write the base64Wrapped content to the part
					e.base64Wrap(ap, a.content)
				}

				if isMixed || isAlternative {
					relatedWriter.Close()
				}
			}
		}
		if isMixed && isAlternative {
			if err := subWriter.Close(); err != nil {
				return nil, err
			}
		}
	}
	// Create attachment part, if necessary
	for _, a := range otherAttachments {
		a.setDefaultHeaders()
		ap, err := w.CreatePart(a.header)
		if err != nil {
			return nil, err
		}
		// Write the base64Wrapped content to the part
		e.base64Wrap(ap, a.content)
	}
	if isMixed || isAlternative || isRelated {
		if err := w.Close(); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (e *Email) headers() (textproto.MIMEHeader, error) {
	res := make(textproto.MIMEHeader, len(e.Headers)+6)
	if e.Headers != nil {
		for _, h := range []string{"Reply-To", "To", "Cc", "From", "Subject", "Date", "Message-Id", "MIME-Version"} {
			if v, ok := e.Headers[h]; ok {
				res[h] = v
			}
		}
	}

	if _, ok := res["Reply-To"]; !ok && len(e.ReplyTo) != 0 {
		res.Set("Reply-to", strings.Join(e.ReplyTo, ", "))
	}
	if _, ok := res["To"]; !ok && len(e.To) != 0 {
		res.Set("To", strings.Join(e.To, ", "))
	}
	if _, ok := res["Cc"]; !ok && len(e.Cc) != 0 {
		res.Set("Cc", strings.Join(e.Cc, ", "))
	}
	if _, ok := res["Message-Id"]; !ok {
		id, err := generateMessageID()
		if err != nil {
			return nil, err
		}
		res.Set("Message-Id", id)
	}
	if _, ok := res["From"]; !ok {
		res.Set("From", e.From)
	}
	if _, ok := res["Date"]; !ok {
		res.Set("Date", time.Now().Format(time.RFC3339))
	}
	if _, ok := res["MIME-Version"]; !ok {
		res.Set("MIME-Version", "1.0")
	}
	for field, vals := range e.Headers {
		if _, ok := res[field]; !ok {
			res[field] = vals
		}
	}
	return res, nil
}

func (e *Email) categorizeAttachments() (htmlRelated, others []*attachment) {
	for _, a := range e.Attachments {
		if a.htmlRelated {
			htmlRelated = append(htmlRelated, a)
		} else {
			others = append(others, a)
		}
	}
	return
}

func (e *Email) headersToBytes(buff io.Writer, header textproto.MIMEHeader) {
	for field, vals := range header {
		for _, subval := range vals {
			io.WriteString(buff, field)
			io.WriteString(buff, ": ")

			switch {
			case field == "Content-Type" || field == "Content-Disposition":
				io.WriteString(buff, subval)
			case field == "From" || field == "To" || field == "Cc" || field == "Bcc":
				participants := strings.Split(subval, ",")
				for i, v := range participants {
					addr, err := mail.ParseAddress(v)
					if err != nil {
						continue
					}
					participants[i] = addr.String()
				}
				io.WriteString(buff, strings.Join(participants, ", "))
			default:
				io.WriteString(buff, mime.QEncoding.Encode("UTF-8", subval))
			}
			io.WriteString(buff, "\r\n")
		}
	}
}

func (e *Email) writeMessage(buf io.Writer, msg []byte, multipart bool, mediaType string, w *multipart.Writer) error {
	if multipart {
		header := textproto.MIMEHeader{
			"Content-Type":              {mediaType + "; charset=UTF-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		}
		if _, err := w.CreatePart(header); err != nil {
			return err
		}
	}

	qp := quotedprintable.NewWriter(buf)
	if _, err := qp.Write(msg); err != nil {
		return err
	}
	return qp.Close()
}

const MaxLineLength = 76

func (e *Email) base64Wrap(w io.Writer, b []byte) {
	// 57 raw bytes per 76-byte base64 line.
	const maxRaw = 57
	// Buffer for each line, including trailing CRLF.
	buffer := make([]byte, MaxLineLength+len("\r\n"))
	copy(buffer[MaxLineLength:], "\r\n")
	// Process raw chunks until there's no longer enough to fill a line.
	for len(b) >= maxRaw {
		base64.StdEncoding.Encode(buffer, b[:maxRaw])
		w.Write(buffer)
		b = b[maxRaw:]
	}
	// Handle the last chunk of bytes.
	if len(b) > 0 {
		out := buffer[:base64.StdEncoding.EncodedLen(len(b))]
		base64.StdEncoding.Encode(out, b)
		out = append(out, "\r\n"...)
		w.Write(out)
	}
}
