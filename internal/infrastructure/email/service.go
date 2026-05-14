package email

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hrms/internal/infrastructure/config"
	"hrms/internal/infrastructure/retry"
	"math/big"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type Service struct {
	client      *sesv2.Client
	senderEmail string
}

func NewService(cfg *config.Config) (*Service, error) {
	if cfg.AWS.Region == "" {
		return nil, fmt.Errorf("AWS region is required for SES")
	}

	var loadOpts []func(*awsconfig.LoadOptions) error

	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		loadOpts = append(loadOpts,
			awsconfig.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					cfg.AWS.AccessKeyID,
					cfg.AWS.SecretAccessKey,
					"",
				),
			),
		)
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		append(loadOpts,
			awsconfig.WithRegion(cfg.AWS.Region),
		)...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for SES: %w", err)
	}

	if cfg.SES.SenderEmail == "" {
		return nil, fmt.Errorf("SES sender email is required (config [ses] sender_email)")
	}

	return &Service{
		client:      sesv2.NewFromConfig(awsCfg),
		senderEmail: cfg.SES.SenderEmail,
	}, nil
}

func (s *Service) GenerateOTP() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(900000))
	return fmt.Sprintf("%06d", n.Int64()+100000)
}

func (s *Service) SendOTP(ctx context.Context, toEmail, otp string) error {
	subject := "Your Verification Code - HRMS Organization Sign-Up"

	bodyText := fmt.Sprintf(`Hello,

Your one-time password (OTP) to complete registration is: %s

This code is valid for 10 minutes.

Do not share this code with anyone.

Best regards,
HRMS System Team`, otp)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.senderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(subject)},
				Body: &types.Body{
					Text: &types.Content{Data: aws.String(bodyText)},
				},
			},
		},
	}

	err := retry.Do(ctx, "SES.SendOTP", func() error {
		_, err := s.client.SendEmail(ctx, input)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to send OTP email to %s: %w", toEmail, err)
	}
	return nil
}

func (s *Service) SendPayslip(ctx context.Context, toEmail, subject, bodyText, attachmentFilename string, attachment []byte) error {
	if len(attachment) == 0 {
		return fmt.Errorf("payslip attachment is empty")
	}
	rawMessage := buildPayslipMIME(s.senderEmail, toEmail, subject, bodyText, attachmentFilename, attachment)
	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(s.senderEmail),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Content: &types.EmailContent{
			Raw: &types.RawMessage{Data: rawMessage},
		},
	}

	err := retry.Do(ctx, "SES.SendPayslip", func() error {
		_, err := s.client.SendEmail(ctx, input)
		return err
	})
	if err != nil {
		return fmt.Errorf("failed to send payslip email to %s: %w", toEmail, err)
	}
	return nil
}

func buildPayslipMIME(fromEmail, toEmail, subject, bodyText, attachmentFilename string, attachment []byte) []byte {
	boundary := "HRMS-PAYSLIP-MIXED-BOUNDARY"
	filename := sanitizeMIMEFilename(attachmentFilename)
	if filename == "" {
		filename = "payslip.pdf"
	}

	var raw bytes.Buffer
	raw.WriteString("From: " + fromEmail + "\r\n")
	raw.WriteString("To: " + toEmail + "\r\n")
	raw.WriteString("Subject: " + subject + "\r\n")
	raw.WriteString("MIME-Version: 1.0\r\n")
	raw.WriteString(`Content-Type: multipart/mixed; boundary="` + boundary + `"` + "\r\n\r\n")

	raw.WriteString("--" + boundary + "\r\n")
	raw.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	raw.WriteString("Content-Transfer-Encoding: 7bit\r\n\r\n")
	raw.WriteString(bodyText + "\r\n\r\n")

	raw.WriteString("--" + boundary + "\r\n")
	raw.WriteString(`Content-Type: application/pdf; name="` + filename + `"` + "\r\n")
	raw.WriteString("Content-Transfer-Encoding: base64\r\n")
	raw.WriteString(`Content-Disposition: attachment; filename="` + filename + `"` + "\r\n\r\n")
	raw.WriteString(wrapBase64(base64.StdEncoding.EncodeToString(attachment)))
	raw.WriteString("\r\n--" + boundary + "--\r\n")

	return raw.Bytes()
}

func sanitizeMIMEFilename(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, `"`, "")
	return value
}

func wrapBase64(value string) string {
	if value == "" {
		return ""
	}
	var b strings.Builder
	for len(value) > 76 {
		b.WriteString(value[:76])
		b.WriteString("\r\n")
		value = value[76:]
	}
	b.WriteString(value)
	return b.String()
}
