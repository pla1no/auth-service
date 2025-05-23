package utils

import (
	"crypto/tls"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"gopkg.in/gomail.v2"
)

type Config struct {
	Host     string
	Port     int
	Address  string
	Password string
}

func SendResetEmail(toEmail, resetURL string) error {
	cfg := Config{
		Host:     viper.GetString("mail.host"),
		Port:     viper.GetInt("mail.port"),
		Address:  viper.GetString("mail.address"),
		Password: viper.GetString("mail.password"),
	}

	m := gomail.NewMessage()
	m.SetHeader("From", "testforserv@mail.ru")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Password reset request")

	plain := fmt.Sprintf(
		"To reset your password, click the link:\n\n%s\n\nThis link will expire in 30 minutes.",
		resetURL,
	)
	m.SetBody("text/plain", plain)

	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Address, cfg.Password)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true, ServerName: cfg.Host}

	conn, err := d.Dial()
	if err != nil {
		log.Printf("SMTP Dial error: %v\n", err)
		return err
	}
	defer conn.Close()

	if err := gomail.Send(conn, m); err != nil {
		log.Printf("SMTP Send error: %v\n", err)
		return err
	}
	return nil
}
