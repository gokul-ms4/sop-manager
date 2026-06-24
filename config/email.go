package config

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"os"
)

type EmailData struct {
	To      string
	Subject string
	Body    string
}

func SendEmail(data EmailData) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	log.Printf("Sending email from %s to %s via %s:%s", from, data.To, smtpHost, smtpPort)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	message := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		from, data.To, data.Subject, data.Body,
	)

	err := smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		[]string{data.To},
		[]byte(message),
	)
	if err != nil {
		log.Printf("Email error: %v", err)
	}
	return err
}

func renderTemplate(templateFile string, data interface{}) (string, error) {
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func WelcomeEmailTemplate(username string, email string) EmailData {
	body, err := renderTemplate("templates/welcome.html", map[string]string{
		"Username": username,
		"Email":    email,
	})
	if err != nil {
		log.Printf("Template error: %v", err)
		body = "Welcome to Tik Talk, " + username
	}

	return EmailData{
		To:      email,
		Subject: "Welcome to Tik Talk! 🎉",
		Body:    body,
	}
}

func PasswordChangedEmailTemplate(username string, email string) EmailData {
	body, err := renderTemplate("templates/password_changed.html", map[string]string{
		"Username": username,
	})
	if err != nil {
		log.Printf("Template error: %v", err)
		body = "Your password was changed, " + username
	}

	return EmailData{
		To:      email,
		Subject: "Your password has been changed",
		Body:    body,
	}
}

func ForgotPasswordEmailTemplate(username string, email string, resetLink string) EmailData {
	body, err := renderTemplate("templates/forgot_password.html", map[string]string{
		"Username":  username,
		"ResetLink": resetLink,
	})
	if err != nil {
		log.Printf("Template error: %v", err)
		body = "Reset your password: " + resetLink
	}

	return EmailData{
		To:      email,
		Subject: "Reset your Tik Talk password",
		Body:    body,
	}
}
