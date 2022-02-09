/*
 *   Corona-Warn-App / cwa-map-backend
 *
 *   (C) 2020, T-Systems International GmbH
 *
 *   Deutsche Telekom AG and all other contributors /
 *   copyright owners license this file to you under the Apache
 *   License, Version 2.0 (the "License"); you may not use this
 *   file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing,
 *   software distributed under the License is distributed on an
 *   "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 *   KIND, either express or implied.  See the License for the
 *   specific language governing permissions and limitations
 *   under the License.
 */

package services

import (
	"context"
	mail "github.com/xhit/go-simple-mail/v2"
	"net/smtp"
	"strings"
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
		AddTo(strings.Split(receiver, ";")...).
		SetSubject(subject).
		SetBody(contentTypeArg, body)

	if email.Error != nil {
		return email.Error
	}

	if err := email.Send(smtpClient); err != nil {
		return err
	}

	return smtpClient.Quit()
}
