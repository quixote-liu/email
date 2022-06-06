package email

import (
	"fmt"
	"net/textproto"
)

type attachment struct {
	filename    string
	contentType string
	header      textproto.MIMEHeader
	content     []byte
	htmlRelated bool
}

func (at *attachment) setDefaultHeaders() {
	contentType := "application/octet-stream"
	if len(at.contentType) > 0 {
		contentType = at.contentType
	}
	at.header.Set("Content-Type", contentType)

	if len(at.header.Get("Content-Disposition")) == 0 {
		disposition := "attachment"
		if at.htmlRelated {
			disposition = "inline"
		}
		at.header.Set("Content-Disposition", fmt.Sprintf("%s;\r\n filename=\"%s\"", disposition, at.filename))
	}
	if len(at.header.Get("Content-ID")) == 0 {
		at.header.Set("Content-ID", fmt.Sprintf("<%s>", at.filename))
	}
	if len(at.header.Get("Content-Transfer-Encoding")) == 0 {
		at.header.Set("Content-Transfer-Encoding", "base64")
	}
}
