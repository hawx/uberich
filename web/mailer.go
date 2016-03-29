package web

import (
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

type Mailer struct {
	addr string
	auth smtp.Auth
	from string
}

func NewMailer(user, pass, addr, from string) *Mailer {
	host := strings.Split(addr, ":")[0]

	auth := smtp.PlainAuth("", user, pass, host)

	return &Mailer{
		addr: addr,
		auth: auth,
		from: from,
	}
}

func (m *Mailer) Send(to, subject, body string) error {
	msg := fmt.Sprintf(`To: %s
Subject: %s

%s`, to, subject, body)

	// return smtp.SendMail(m.addr, m.auth, m.from, []string{to}, []byte(msg))

	log.Println(msg)
	return nil
}
