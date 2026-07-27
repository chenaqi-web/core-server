package mail

import (
	"context"
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"backend/core-server/internal/config"
	"backend/core-server/internal/domain"
)

type sentMail struct {
	address string
	from    string
	to      []string
	message string
}

func TestQQMailSenderUsesInjectedSMTPTransport(t *testing.T) {
	var captured sentMail
	fakeSend := func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
		captured = sentMail{address: addr, from: from, to: to, message: string(msg)}
		return nil
	}
	sender, err := newQQMailSender(&config.Config{Auth: config.AuthConfig{
		QQSMTPUsername: "sender@qq.com",
		QQSMTPAuthCode: "fake-auth-code",
	}}, fakeSend)
	if err != nil {
		t.Fatalf("newQQMailSender() error = %v", err)
	}

	if err := sender.SendVerificationCode(
		context.Background(),
		"recipient@qq.com",
		"123456",
		domain.EmailCodePurposeRegister,
	); err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}
	if captured.address != qqSMTPAddress || captured.from != "sender@qq.com" {
		t.Fatalf("unexpected SMTP envelope: %+v", captured)
	}
	if len(captured.to) != 1 || captured.to[0] != "recipient@qq.com" {
		t.Fatalf("unexpected recipients: %v", captured.to)
	}
	if !strings.Contains(captured.message, "123456") || !strings.Contains(captured.message, "register") {
		t.Fatalf("message does not contain expected content: %q", captured.message)
	}
	if strings.Contains(captured.message, "fake-auth-code") {
		t.Fatal("SMTP authorization code leaked into the message")
	}
}

func TestQQMailSenderDoesNotSendInvalidRequest(t *testing.T) {
	called := false
	sender, err := newQQMailSender(&config.Config{Auth: config.AuthConfig{
		QQSMTPUsername: "sender@qq.com",
		QQSMTPAuthCode: "fake-auth-code",
	}}, func(string, smtp.Auth, string, []string, []byte) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("newQQMailSender() error = %v", err)
	}

	err = sender.SendVerificationCode(context.Background(), "recipient@qq.com", "invalid", domain.EmailCodePurposeRegister)
	if !errors.Is(err, ErrInvalidMail) {
		t.Fatalf("SendVerificationCode() error = %v, want %v", err, ErrInvalidMail)
	}
	if called {
		t.Fatal("fake SMTP transport was called for an invalid request")
	}
}

func TestQQMailSenderHonorsCanceledContext(t *testing.T) {
	called := false
	sender, err := newQQMailSender(&config.Config{Auth: config.AuthConfig{
		QQSMTPUsername: "sender@qq.com",
		QQSMTPAuthCode: "fake-auth-code",
	}}, func(string, smtp.Auth, string, []string, []byte) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("newQQMailSender() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := sender.SendVerificationCode(ctx, "recipient@qq.com", "123456", domain.EmailCodePurposeRegister); !errors.Is(err, context.Canceled) {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}
	if called {
		t.Fatal("fake SMTP transport was called for a canceled context")
	}
}

func TestNewQQMailSenderRequiresEnvironmentConfig(t *testing.T) {
	if _, err := newQQMailSender(&config.Config{}, func(string, smtp.Auth, string, []string, []byte) error { return nil }); err == nil {
		t.Fatal("newQQMailSender() expected a configuration error")
	}
}
