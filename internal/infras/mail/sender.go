package mail

import (
	"fmt"
	"net"
	"net/smtp"
	"strconv"

	"core-server/internal/config"
)

type Sender struct {
	cfg config.EmailConfig
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg.Email}
}

// 发送邮箱验证码
func (s *Sender) SendCode(to, code string) error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	auth := smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	message := []byte(fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: Email verification code\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nYour verification code is %s. It is valid for 5 minutes.\r\n",
		to,
		s.cfg.From,
		code,
	))
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, message)
}
