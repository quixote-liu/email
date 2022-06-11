package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/textproto"
)

type attachment struct {
	filename    string
	contentType string
	header      textproto.MIMEHeader
	content     *bytes.Buffer
}

func (at *attachment) setDefaultHeader() {
	if at.header == nil {
		at.header = make(textproto.MIMEHeader)
	}

	// set content-type.
	ct := "application/octet-stream"
	if at.contentType != "" {
		ct = at.contentType
	}
	at.header.Set("Content-Type", ct)

	// set disposition
	at.header.Set("Content-Disposition", fmt.Sprintf("attachment; filename='%s'", at.filename))

	// set content-id
	id, _ := generateContentID(at.filename)
	at.header.Set("Content-ID", id)

	// set Content-Transfer-Encoding as base64.
	at.header.Set("Content-Transfer-Encoding", "base64")
}

func (at *attachment) writeBase64To(w io.Writer) {
	src := at.content.Bytes()
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(src)))
	base64.StdEncoding.Encode(dst, src)
	dst = append(dst, "\r\n"...)
	_, _ = w.Write(dst)
}
