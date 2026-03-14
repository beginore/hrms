package cognito

import (
	"context"
	"fmt"
	"hrms/internal/infrastructure/config"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
)

type Client struct {
	svc             *cognitoidentityprovider.Client
	userPoolID      string
	appClientID     string
	appClientSecret string
}

func New(cfg *config.Config) (*Client, error) {
	if cfg.AWS.Region == "" {
		return nil, fmt.Errorf("AWS region is required")
	}

	var loadOpts []func(*awsconfig.LoadOptions) error

	if cfg.AWS.AccessKeyID != "" && cfg.AWS.SecretAccessKey != "" {
		loadOpts = append(loadOpts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AWS.AccessKeyID,
				cfg.AWS.SecretAccessKey,
				"",
			),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		append(loadOpts, awsconfig.WithRegion(cfg.AWS.Region))...,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &Client{
		svc:             cognitoidentityprovider.NewFromConfig(awsCfg),
		userPoolID:      cfg.Cognito.UserPoolID,
		appClientID:     cfg.Cognito.AppClientID,
		appClientSecret: cfg.Cognito.AppClientSecret,
	}, nil
}

func (c *Client) PoolID() string                       { return c.userPoolID }
func (c *Client) AppClientID() string                  { return c.appClientID }
func (c *Client) AppClientSecret() string              { return c.appClientSecret }
func (c *Client) Svc() *cognitoidentityprovider.Client { return c.svc }
