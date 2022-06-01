package email

import "net/textproto"

type attachment struct {
	filename    string
	contentType string
	header      textproto.MIMEHeader
	content     []byte
	htmlRelated bool
}

func (at *attachment) setDefaultHeaders() {
	// TODO: optimize.
}
