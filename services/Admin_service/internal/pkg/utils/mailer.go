package utils

import (
	"fmt"
	"net/smtp"
)

type Mailer struct {
	host string
	port string
	user string
	pass string
}

func NewMailer(host, port, user, pass string) *Mailer {
	return &Mailer{
		host: host,
		port: port,
		user: user,
		pass: pass,
	}
}

func (m *Mailer) SendOtp(email, otp string) error {

	auth := smtp.PlainAuth("", m.user, m.pass, m.host)

	msg := []byte(fmt.Sprintf(
		"Subject: Password Reset OTP\n\nYour OTP is: %s\nValid for 5 minutes.",
		otp,
	))

	addr := m.host + ":" + m.port

	return smtp.SendMail(
		addr,
		auth,
		m.user,
		[]string{email},
		msg,
	)
}