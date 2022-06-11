# email
一个可以发送邮件工具。发送普通邮件示例:
```golang
    func main() {
        e := email.Email{
            From:     "your_email@mail.com",
            To:       []string{"target_email@qq.com"},
            Subject:  "This is Subject",
            Addr:     "smtp.example.com:25",
        }

        if err := e.AttachFile("./filename.txt"); err != nil {
        	log.Fatal(err)
        }

        auth := smtp.PlainAuth("", "your_email@mail.com", "your_password", "smtp.example.com")
        message := []byte("hello, world")

        err := e.SetAuth(auth).WriteText(message).Send()
        if err != nil {
            log.Fatal(err)
        }

        // reset email message, include attchments.
        e.Reset()
    }
```

参考github.com/jordan-wright/email和smtp协议写的一个可以发送邮件的工具，主要以学习为目的。