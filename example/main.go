package main

import (
	"email"
	"log"
	"net/smtp"
)

func main() {
	// SendText()
	SendHTML()
	// Send()
}

func SendText() {
	e := email.Email{
		From:     "lcs.shun@foxmail.com",
		To:       []string{"1351169665@qq.com"},
		Subject:  "email testing",
		Comments: "this is comments",
		Addr:     "smtp.qq.com:25",
	}
	auth := smtp.PlainAuth("", "lcs.shun@foxmail.com", "ikpfnntluodtbadh", "smtp.qq.com")

	text := `可以看出，如果在邮件中要添加附件，必须定义multipart/mixed段；如果存在内嵌资源，至少要定义multipart/related段；
	如果纯文本与超文本共存，至少要定义multipart/alternative段。什么是“至少”？举个例子说，如果只有纯文本与超文本正文，
	那么在邮件头中将类型扩大化，定义为multipart/related，`
	e.SetAuth(auth).WriteText([]byte(text))

	if err := e.AttachFile("./shenghuo.txt"); err != nil {
		log.Fatal(err)
	}
	// if err := e.AttachFile("./photo.png"); err != nil {
	// 	log.Fatal(err)
	// }

	err := e.Send()
	if err != nil {
		log.Fatal(err)
	}

	e.Reset()
}

func SendHTML() {
	e := email.Email{
		From:     "lcs.shun@foxmail.com",
		To:       []string{"1351169665@qq.com"},
		Subject:  "email testing",
		Comments: "this is comments",
		Addr:     "smtp.qq.com:25",
	}
	auth := smtp.PlainAuth("", "lcs.shun@foxmail.com", "ikpfnntluodtbadh", "smtp.qq.com")

	text := `<h1>testing email !</h1>`
	e.SetAuth(auth).WriteHTML([]byte(text))

	if err := e.AttachFile("./shenghuo.txt"); err != nil {
		log.Fatal(err)
	}

	err := e.Send()
	if err != nil {
		log.Fatal(err)
	}

	e.Reset()
}

func Send() {
	message := `Subject: github remote branch
	Message-Id: <1654937364018546523.55472.1852404614153174204@ubuntu>
	From: <1136089132@qq.com>
	Date: Sat, 11 Jun 2022 16:49:24 +0800
	Mime-Version: 1.0
	Content-Type: text/html; charset=UTF-8
	Content-Transfer-Encoding: quoted-printable
	To: <1351169665@qq.com>
	
	<h1>Fancy HTML is supported, too!</h1>`

	auth := smtp.PlainAuth("", "lcs.shun@foxmail.com", "ikpfnntluodtbadh", "smtp.qq.com")

	err := smtp.SendMail("smtp.qq.com:25", auth, "lcs.shun@foxmail.com", []string{"1351169665@qq.com"}, []byte(message))
	if err != nil {
		log.Fatal(err)
	}
}
