package emails

import (
	"fmt"
	"os"

	"gopkg.in/gomail.v2"
)

func RemoveEmail(userEmail, tournamentName string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetHeader("To", userEmail)
	text := fmt.Sprintf("Hello, you have recieved this automated email to tell "+
		"you that you have been removed from tournament %s. If you think that this has been a "+
		"mistake please contact the tournament owner", tournamentName)
	subject := fmt.Sprintf("Removing from tournament %s", tournamentName)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", text)

	d := gomail.NewDialer(os.Getenv("SMTP_FROM"), 465, os.Getenv("SMTP_EMAIL"), os.Getenv("SMTP_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

func NotifyChangeEmail(emails []string, tournamentName string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_FROM"))
	m.SetHeader("To", os.Getenv("SMTP_FROM"))
	m.SetHeader("Bcc", emails...)
	text := fmt.Sprintf("Hello, you have recieved this automated email to tell "+
		"you that you have been removed from tournament %s. If you think that this has been a "+
		"mistake please contact the tournament owner", tournamentName)
	subject := fmt.Sprintf("Removing from tournament %s", tournamentName)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", text)

	d := gomail.NewDialer(os.Getenv("SMTP_FROM"), 465, os.Getenv("SMTP_EMAIL"), os.Getenv("SMTP_PASSWORD"))

	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
