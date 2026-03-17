package service

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"hrms/internal/infrastructure/config"
)

type mailer struct {
	host        string
	port        int
	username    string
	password    string
	senderEmail string
	enabled     bool
}

func newMailer(cfg *config.Config) (*mailer, error) {
	senderEmail := strings.TrimSpace(cfg.SMTP.SenderEmail)
	if senderEmail == "" {
		senderEmail = strings.TrimSpace(cfg.SMTP.Username)
	}

	enabled := strings.TrimSpace(cfg.SMTP.Host) != "" &&
		cfg.SMTP.Port > 0 &&
		strings.TrimSpace(cfg.SMTP.Username) != "" &&
		strings.TrimSpace(cfg.SMTP.Password) != "" &&
		senderEmail != ""

	return &mailer{
		host:        strings.TrimSpace(cfg.SMTP.Host),
		port:        cfg.SMTP.Port,
		username:    strings.TrimSpace(cfg.SMTP.Username),
		password:    cfg.SMTP.Password,
		senderEmail: senderEmail,
		enabled:     enabled,
	}, nil
}

func (m *mailer) SendInvite(ctx context.Context, toEmail, firstName, organizationName, inviteCode, platformURL string) error {
	_ = ctx
	if !m.enabled {
		log.Printf("[Invite Mailer] SMTP delivery is not configured for email=%q", toEmail)
		return ErrInviteEmailUnavailable
	}

	log.Printf("[Invite Mailer] Sending SMTP invite email to=%q via host=%q port=%d", toEmail, m.host, m.port)

	subject := "You have been invited to join HRMS"

	bodyText := fmt.Sprintf(`Hello %s,

Your company administrator invited you to join the HRMS platform for %s.

Invitation Code: %s

Join the platform here: %s

Enter this code on the registration page to activate your account.

This invitation expires in 24 hours and can only be used once.

Best regards,
HRMS System Team`, firstName, organizationName, inviteCode, platformURL)

	message := strings.Join([]string{
		fmt.Sprintf("From: %s", m.senderEmail),
		fmt.Sprintf("To: %s", toEmail),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=\"UTF-8\"",
		"",
		bodyText,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	err := smtp.SendMail(addr, auth, m.senderEmail, []string{toEmail}, []byte(message))
	if err != nil {
		log.Printf("[Invite Mailer] SMTP send failed to=%q: %v", toEmail, err)
		return fmt.Errorf("failed to send invite email to %s via smtp: %w", toEmail, err)
	}

	log.Printf("[Invite Mailer] SMTP invite email sent to=%q", toEmail)

	return nil
}
