package email

import (
	"context"
	"crypto/rand"
	"fmt"
	"hrms/internal/infrastructure/config"
	"math/big"

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

	_, err := s.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send OTP email to %s: %w", toEmail, err)
	}

	return nil
}
