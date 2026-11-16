package email

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"modl-go/internal/config"
)

type Sender struct {
	cfg *config.Config
}

func NewSender(cfg *config.Config) *Sender {
	return &Sender{cfg: cfg}
}

func (s *Sender) IsConfigured() bool {
	return strings.TrimSpace(s.cfg.SMTPHost) != ""
}

func (s *Sender) Send(to, subject, htmlBody string) error {
	if !s.IsConfigured() {
		log.Printf("[EMAIL-FALLBACK] To=%s | Subject=%s | Body=%s", to, subject, stripTags(htmlBody))
		return nil
	}
	from := fmt.Sprintf("noreply@%s", s.cfg.EmailDomain)
	if s.cfg.SMTPUsername != "" && strings.Contains(s.cfg.SMTPUsername, "@") {
		from = s.cfg.SMTPUsername
	}
	addr := fmt.Sprintf("%s:%d", s.cfg.SMTPHost, s.cfg.SMTPPort)
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s <%s>", s.cfg.EmailDomain, from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	var auth smtp.Auth
	if s.cfg.SMTPAuth && s.cfg.SMTPUsername != "" {
		auth = smtp.PlainAuth("", s.cfg.SMTPUsername, s.cfg.SMTPPassword, s.cfg.SMTPHost)
	}
	if s.cfg.SMTPTLS {
		c, err := smtp.Dial(addr)
		if err != nil {
			return err
		}
		defer c.Close()
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: s.cfg.SMTPHost, MinVersion: tls.VersionTLS12}); err != nil {
				return err
			}
		}
		if auth != nil {
			if err := c.Auth(auth); err != nil {
				return err
			}
		}
		if err := c.Mail(from); err != nil {
			return err
		}
		if err := c.Rcpt(to); err != nil {
			return err
		}
		w, err := c.Data()
		if err != nil {
			return err
		}
		if _, err := w.Write([]byte(msg)); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
		return c.Quit()
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func (s *Sender) SendAuthCode(to, serverName, code string) error {
	body := fmt.Sprintf("<p>Your login verification code for <b>%s</b> is:</p><h2>%s</h2><p>This code expires in 5 minutes.</p>", serverName, code)
	return s.Send(to, fmt.Sprintf("[%s] Verification code: %s", serverName, code), body)
}

func (s *Sender) SendInvitation(to, serverName, inviteURL string) error {
	body := fmt.Sprintf("<p>You have been invited to join the staff team on <b>%s</b>.</p><p><a href=\"%s\">Accept Invitation</a></p>", serverName, inviteURL)
	return s.Send(to, fmt.Sprintf("[%s] Staff Invitation", serverName), body)
}

func (s *Sender) SendTicketVerification(to, serverName, code string) error {
	body := fmt.Sprintf("<p>Your ticket verification code for <b>%s</b> is:</p><h2>%s</h2>", serverName, code)
	return s.Send(to, fmt.Sprintf("[%s] Ticket Verification Code: %s", serverName, code), body)
}

func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
			b.WriteByte(' ')
		default:
			if !inTag {
				b.WriteRune(r)
			}
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
