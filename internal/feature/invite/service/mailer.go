package service

import (
	"context"
	"fmt"

	"hrms/internal/infrastructure/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type mailer struct {
	client      *sesv2.Client
	senderEmail string
}

func newMailer(cfg *config.Config) (*mailer, error) {
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

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		append(loadOpts, awsconfig.WithRegion(cfg.AWS.Region))...,
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config for invite mailer: %w", err)
	}

	return &mailer{
		client:      sesv2.NewFromConfig(awsCfg),
		senderEmail: cfg.SES.SenderEmail,
	}, nil
}

func (m *mailer) SendInvite(ctx context.Context, toEmail, firstName, organizationName, inviteCode, platformURL string) error {
	subject := "You have been invited to join HRMS"

	bodyText := fmt.Sprintf(`Hello %s,

Your company administrator invited you to join the HRMS platform for %s.

Invitation Code: %s

Join the platform here: %s

Enter this code on the registration page to activate your account.

This invitation expires in 24 hours and can only be used once.

Best regards,
HRMS System Team`, firstName, organizationName, inviteCode, platformURL)

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(m.senderEmail),
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

	_, err := m.client.SendEmail(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to send invite email to %s: %w", toEmail, err)
	}

	return nil
}
