package main

import (
	"email"
	"log"
)

func main() {
	e := email.Email{
		From:    "1136089132@qq.com",
		To:      []string{"1351169665@qq.com"},
		Addr:    "smtp.qq.com:25",
		Subject: "hello",
	}
	if err := e.Send("hello, world"); err != nil {
		log.Printf("send email failed: %v", err)
	}
}
