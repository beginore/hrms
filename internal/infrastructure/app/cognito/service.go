package cognito

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/google/uuid"
)

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserNotConfirmed       = errors.New("user is not confirmed")
	ErrUserAlreadyConfirmed   = errors.New("user is already confirmed")
	ErrInvalidOTPCode         = errors.New("invalid or incorrect OTP code")
	ErrOTPCodeExpired         = errors.New("OTP code has expired")
	ErrInvalidRefreshToken    = errors.New("invalid or expired refresh token")
	ErrCognitoOperationFailed = errors.New("cognito operation failed")
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
		var notAuth *types.NotAuthorizedException
		if errors.As(err, &notAuth) {
			return ErrInvalidCredentials
		}
		var codeMismatch *types.CodeMismatchException
		if errors.As(err, &codeMismatch) {
			return ErrInvalidOTPCode
		}
		var expiredCode *types.ExpiredCodeException
		if errors.As(err, &expiredCode) {
			return ErrOTPCodeExpired
		}
		var alreadyConfirmed *types.NotAuthorizedException
		if errors.As(err, &alreadyConfirmed) {
			return ErrUserAlreadyConfirmed
		}
		var invalidParam *types.InvalidParameterException
		if errors.As(err, &invalidParam) {
			return ErrInvalidOTPCode
		}
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
		var notConfirmed *types.UserNotConfirmedException
		if errors.As(err, &notConfirmed) {
			return nil, ErrUserNotConfirmed
		}
		var notAuth *types.NotAuthorizedException
		if errors.As(err, &notAuth) {
			return nil, ErrInvalidCredentials
		}
		var userNotFound *types.UserNotFoundException
		if errors.As(err, &userNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("%w: %v", ErrCognitoOperationFailed, err)
	}
	return output, nil
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken, username string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	output, err := s.client.Svc().InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		ClientId: aws.String(s.client.AppClientID()),
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": refreshToken,
			"SECRET_HASH":   s.computeSecretHash(username),
		},
	})
	if err != nil {
		log.Printf("[Cognito RefreshTokens] raw AWS error: %T: %v", err, err)
		var notAuth *types.NotAuthorizedException
		if errors.As(err, &notAuth) {
			return nil, ErrInvalidRefreshToken
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

func (s *Service) ParseTokenClaims(accessToken string) (userID uuid.UUID, cognitoUsername string, err error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return uuid.Nil, "", fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to decode token payload: %w", err)
	}

	var claims struct {
		Sub      string `json:"sub"`
		Username string `json:"username"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return uuid.Nil, "", fmt.Errorf("failed to unmarshal token claims: %w", err)
	}

	userID, err = uuid.Parse(claims.Sub)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid sub claim: %w", err)
	}

	return userID, claims.Username, nil
}
