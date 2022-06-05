package main

import (
	"email"
	"log"
	"net/smtp"
	"net/textproto"
)

func main() {
	SendHTML()
}

func SendText() {
	text := `当你想要公开分享一个分支时，需要将其推送到有写入权限的远程仓库上。 
	本地的分支并不会自动与远程仓库同步——你必须显式地推送想要分享的分支。 这样，
	你就可以把不愿意分享的内容放到私人分支上，而将需要和别人协作的内容推送到公开分支。
	如果希望和别人一起在名为 serverfix 的分支上工作，你可以像推送第一个分支那样推送它。
	运行 git push <remote> <branch>:`

	e := email.Email{
		From:    "lcs.shun@foxmail.com",
		To:      []string{"1351169665@qq.com"},
		Subject: "github remote branch",
		Text:    []byte(text),
		// HTML:    []byte("<h1>Fancy HTML is supported, too!</h1>"),
		Headers: textproto.MIMEHeader{},
	}
	e.AttachFile("")
	err := e.Send("smtp.qq.com:25", smtp.PlainAuth("", "lcs.shun@foxmail.com", "ikpfnntluodtbadh", "smtp.qq.com"))
	if err != nil {
		log.Fatal(err)
	}
}

func SendHTML() {
	e := email.Email{
		From:    "lcs.shun@foxmail.com",
		To:      []string{"1351169665@qq.com"},
		Subject: "test",
		HTML:    []byte("<h1>Fancy HTML is supported, too!</h1>"),
		Headers: textproto.MIMEHeader{},
	}
	if err := e.AttachFile("./shenghuo"); err != nil {
		log.Fatal(err)
	}
	err := e.Send("smtp.qq.com:25", smtp.PlainAuth("", "lcs.shun@foxmail.com", "ikpfnntluodtbadh", "smtp.qq.com"))
	if err != nil {
		log.Fatal(err)
	}
}
