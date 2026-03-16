package cognito

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrCognitoOperationFailed = errors.New("cognito operation failed")
	ErrUserNotConfirmed       = errors.New("user account is not confirmed")
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) SignUpUser(ctx context.Context, email, password, firstName, lastName, phone string) (string, error) {
	fullName := firstName + " " + lastName

	input := &cognitoidentityprovider.SignUpInput{
		ClientId:   aws.String(s.client.AppClientID()),
		Username:   aws.String(email),
		Password:   aws.String(password),
		SecretHash: aws.String(s.computeSecretHash(email)),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("name"), Value: aws.String(fullName)},
			{Name: aws.String("phone_number"), Value: aws.String(phone)},
		},
	}

	output, err := s.client.Svc().SignUp(ctx, input)
	if err != nil {
		var exists *types.UsernameExistsException
		if errors.As(err, &exists) {
			return "", ErrUserAlreadyExists
		}
		return "", fmt.Errorf("%w: %v", ErrCognitoOperationFailed, err)
	}

	return aws.ToString(output.UserSub), nil
}

func (s *Service) ConfirmSignUp(ctx context.Context, email, code string) error {
	_, err := s.client.Svc().ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         aws.String(s.client.AppClientID()),
		Username:         aws.String(email),
		ConfirmationCode: aws.String(code),
		SecretHash:       aws.String(s.computeSecretHash(email)),
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCognitoOperationFailed, err)
	}
	return nil
}

func (s *Service) SignIn(ctx context.Context, email, password string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	output, err := s.client.Svc().InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(s.client.AppClientID()),
		AuthParameters: map[string]string{
			"USERNAME":    email,
			"PASSWORD":    password,
			"SECRET_HASH": s.computeSecretHash(email),
		},
	})
	if err != nil {
		var notAuth *types.NotAuthorizedException
		if errors.As(err, &notAuth) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: %v", ErrCognitoOperationFailed, err)
	}
	return output, nil
}

func (s *Service) computeSecretHash(username string) string {
	if s.client.appClientSecret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.client.appClientSecret))
	mac.Write([]byte(username + s.client.appClientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
