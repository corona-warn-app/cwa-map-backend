package services

import (
	"context"
	mail "github.com/xhit/go-simple-mail/v2"
	"net/smtp"
)

type EmailConfig struct {
	SmtpHost     string
	SmtpPort     int
	SmtpUser     string
	SmtpPassword string
	From         string
}

type MailService interface {
	// SendMail send the mail with the given receiver, subject and body to the configured mail server
	SendMail(ctx context.Context, receiver, subject, contentType, body string) error
}

type mailService struct {
	config EmailConfig
	auth   smtp.Auth
}

func NewMailService(config EmailConfig) MailService {
	return &mailService{
		config: config,
		auth:   smtp.PlainAuth("", config.SmtpUser, config.SmtpPassword, config.SmtpHost),
	}
}

func (m *mailService) SendMail(ctx context.Context, receiver, subject, contentType, body string) error {
	server := mail.NewSMTPClient()
	server.Host = m.config.SmtpHost
	server.Port = m.config.SmtpPort
	server.Username = m.config.SmtpUser
	server.Password = m.config.SmtpPassword
	server.Encryption = mail.EncryptionSTARTTLS
	server.Authentication = mail.AuthLogin
	server.KeepAlive = true

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	contentTypeArg := mail.TextPlain
	if contentType == "text/html" {
		contentTypeArg = mail.TextHTML
	}

	email := mail.NewMSG()
	email.SetFrom(m.config.From).
		AddTo(receiver).
		SetSubject(subject).
		SetBody(contentTypeArg, body)

	if email.Error != nil {
		return err
	}

	if err := email.Send(smtpClient); err != nil {
		return err
	}

	return smtpClient.Quit()
}
