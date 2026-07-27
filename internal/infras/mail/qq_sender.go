package mail

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
)

const (
	qqSMTPHost    = "smtp.qq.com"
	qqSMTPAddress = qqSMTPHost + ":587"
)

var ErrInvalidMail = errors.New("invalid mail request")

type sendMailFunc func(addr string, auth smtp.Auth, from string, to []string, msg []byte) error

type QQMailSender struct {
	username string
	auth     smtp.Auth
	sendMail sendMailFunc
}

var _ domain.MailSender = (*QQMailSender)(nil)

func NewQQMailSender(cfg *config.Config) (*QQMailSender, error) {
	return newQQMailSender(cfg, smtp.SendMail)
}

func newQQMailSender(cfg *config.Config, sendMail sendMailFunc) (*QQMailSender, error) {
	if cfg == nil {
		return nil, fmt.Errorf("QQ SMTP config is nil")
	}
	username := strings.TrimSpace(cfg.Auth.QQSMTPUsername)
	authCode := strings.TrimSpace(cfg.Auth.QQSMTPAuthCode)
	if username == "" || authCode == "" {
		return nil, fmt.Errorf("QQ SMTP environment variables are required")
	}
	if sendMail == nil {
		return nil, fmt.Errorf("SMTP sender is nil")
	}
	return &QQMailSender{
		username: username,
		auth:     smtp.PlainAuth("", username, authCode, qqSMTPHost),
		sendMail: sendMail,
	}, nil
}

func (s *QQMailSender) SendVerificationCode(
	ctx context.Context,
	recipient, code string,
	purpose domain.EmailCodePurpose,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	recipient = strings.TrimSpace(recipient)
	if recipient == "" || strings.ContainsAny(recipient, "\r\n") || !isMailCode(code) || !purpose.Valid() {
		return ErrInvalidMail
	}

	message := []byte(strings.Join([]string{
		"From: " + s.username,
		"To: " + recipient,
		"Subject: Verification code",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		fmt.Sprintf("Your %s verification code is %s. It expires in 5 minutes.", purpose, code),
	}, "\r\n"))

	if err := s.sendMail(qqSMTPAddress, s.auth, s.username, []string{recipient}, message); err != nil {
		return fmt.Errorf("send QQ SMTP mail: %w", err)
	}
	return nil
}

func isMailCode(code string) bool {
	if len(code) != 6 {
		return false
	}
	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
