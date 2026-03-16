package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	"hrms/internal/feature/invite/repository"
	"hrms/internal/infrastructure/app/cognito"
	"hrms/internal/infrastructure/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	cognitoTypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/google/uuid"
)

const (
	inviteTTL         = 24 * time.Hour
	maxInviteAttempts = 10
	defaultPortalURL  = "http://localhost:3000/register"
	defaultInviteRole = "Employee"
)

type Service struct {
	repo          *repository.Repository
	cognitoClient *cognito.Client
	mailer        *mailer
}

func NewService(repo *repository.Repository, cfg *config.Config, cognitoClient *cognito.Client) (*Service, error) {
	inviteMailer, err := newMailer(cfg)
	if err != nil {
		return nil, err
	}

	return &Service{
		repo:          repo,
		cognitoClient: cognitoClient,
		mailer:        inviteMailer,
	}, nil
}

func (s *Service) GenerateInvite(ctx context.Context, req GenerateInviteRequest) (*GenerateInviteResponse, error) {
	orgIDRaw := strings.TrimSpace(req.OrganizationID)
	if orgIDRaw == "" {
		return nil, ErrOrganizationIDRequired
	}
	if strings.TrimSpace(req.FirstName) == "" {
		return nil, ErrFirstNameRequired
	}
	if strings.TrimSpace(req.LastName) == "" {
		return nil, ErrLastNameRequired
	}
	if strings.TrimSpace(req.Email) == "" {
		return nil, ErrEmailRequired
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(req.Email)); err != nil {
		return nil, ErrInvalidEmail
	}

	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		return nil, ErrInvalidOrganizationID
	}

	organizationName, err := s.repo.GetOrganizationNameByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrganizationNotFound
		}
		return nil, fmt.Errorf("get organization by id: %w", err)
	}

	inviteID := uuid.New()
	expiresAt := time.Now().UTC().Add(inviteTTL)

	var code string
	for range maxInviteAttempts {
		code, err = generateInviteCode()
		if err != nil {
			return nil, fmt.Errorf("generate invite code: %w", err)
		}

		insertErr := s.repo.CreateInvite(ctx, repository.CreateInviteParams{
			ID:        inviteID,
			OrgID:     orgID,
			FirstName: strings.TrimSpace(req.FirstName),
			LastName:  strings.TrimSpace(req.LastName),
			Email:     strings.TrimSpace(req.Email),
			Code:      code,
			Role:      resolveRole(req.Role),
			Position:  trimOptional(req.Position),
			ExpiresAt: expiresAt,
		})
		if insertErr == nil {
			err = nil
			break
		}
		if !repository.IsUniqueViolation(insertErr) {
			return nil, fmt.Errorf("create invite: %w", insertErr)
		}
		err = insertErr
	}
	if err != nil {
		return nil, ErrGenerateInvite
	}

	if err := s.mailer.SendInvite(
		ctx,
		strings.TrimSpace(req.Email),
		strings.TrimSpace(req.FirstName),
		organizationName,
		code,
		defaultPortalURL,
	); err != nil {
		_ = s.repo.DeleteInviteByID(ctx, inviteID)
		return nil, fmt.Errorf("send invite email: %w", err)
	}

	return &GenerateInviteResponse{
		InviteID:         inviteID.String(),
		OrganizationID:   orgID.String(),
		OrganizationName: organizationName,
		Email:            strings.TrimSpace(req.Email),
		Code:             code,
		ExpiresAt:        expiresAt,
	}, nil
}

func (s *Service) VerifyInvite(ctx context.Context, req VerifyInviteRequest) (*VerifyInviteResponse, error) {
	code := normalizeInviteCode(req.Code)
	if code == "" {
		return nil, ErrInviteCodeRequired
	}

	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInviteNotFound
		}
		return nil, fmt.Errorf("get invite by code: %w", err)
	}

	if err := validateInvite(invite); err != nil {
		return nil, err
	}

	return &VerifyInviteResponse{
		OrganizationID:   invite.OrgID,
		OrganizationName: invite.OrganizationName,
		FirstName:        invite.FirstName,
		LastName:         invite.LastName,
		Email:            invite.Email,
		Role:             invite.Role,
		Position:         invite.Position,
		ExpiresAt:        invite.ExpiresAt,
		Message:          fmt.Sprintf("You have been invited to join %s", invite.OrganizationName),
	}, nil
}

func (s *Service) CompleteRegistration(ctx context.Context, req CompleteRegistrationRequest) (*CompleteRegistrationResponse, error) {
	code := normalizeInviteCode(req.Code)
	if code == "" {
		return nil, ErrInviteCodeRequired
	}
	if !isValidPassword(req.Password) {
		return nil, ErrPasswordRequired
	}
	if strings.TrimSpace(req.PhoneNumber) == "" {
		return nil, ErrPhoneNumberRequired
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	invite, err := s.repo.GetInviteByCodeTx(ctx, tx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInviteNotFound
		}
		return nil, fmt.Errorf("lock invite by code: %w", err)
	}

	if err := validateInvite(invite); err != nil {
		return nil, err
	}

	firstName := strings.TrimSpace(req.FirstName)
	if firstName == "" {
		firstName = invite.FirstName
	}
	lastName := strings.TrimSpace(req.LastName)
	if lastName == "" {
		lastName = invite.LastName
	}

	userSub, err := s.createConfirmedUser(
		ctx,
		invite.Email,
		req.Password,
		firstName,
		lastName,
		strings.TrimSpace(req.PhoneNumber),
	)
	if err != nil {
		if errors.Is(err, cognito.ErrUserAlreadyExists) {
			return nil, ErrUserAlreadyExists
		}
		return nil, fmt.Errorf("create cognito user: %w", err)
	}

	userID, err := uuid.Parse(userSub)
	if err != nil {
		return nil, fmt.Errorf("parse cognito user id: %w", err)
	}
	orgID, err := uuid.Parse(invite.OrgID)
	if err != nil {
		return nil, fmt.Errorf("parse org id: %w", err)
	}
	inviteID, err := uuid.Parse(invite.ID)
	if err != nil {
		return nil, fmt.Errorf("parse invite id: %w", err)
	}

	if err := s.repo.InsertUserTx(ctx, tx, repository.CreateInvitedUserParams{
		ID:          userID,
		OrgID:       orgID,
		Email:       invite.Email,
		Role:        invite.Role,
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: strings.TrimSpace(req.PhoneNumber),
	}); err != nil {
		return nil, fmt.Errorf("insert invited user: %w", err)
	}

	if err := s.repo.MarkInviteUsedTx(ctx, tx, inviteID, time.Now().UTC()); err != nil {
		return nil, fmt.Errorf("mark invite used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit registration: %w", err)
	}

	return &CompleteRegistrationResponse{
		UserID:         userID.String(),
		OrganizationID: invite.OrgID,
		Role:           invite.Role,
	}, nil
}

func (s *Service) createConfirmedUser(ctx context.Context, email, password, firstName, lastName, phone string) (string, error) {
	output, err := s.cognitoClient.Svc().AdminCreateUser(ctx, &cognitoidentityprovider.AdminCreateUserInput{
		UserPoolId:    aws.String(s.cognitoClient.PoolID()),
		Username:      aws.String(email),
		MessageAction: cognitoTypes.MessageActionTypeSuppress,
		UserAttributes: []cognitoTypes.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("given_name"), Value: aws.String(strings.TrimSpace(firstName))},
			{Name: aws.String("family_name"), Value: aws.String(strings.TrimSpace(lastName))},
			{Name: aws.String("name"), Value: aws.String(strings.TrimSpace(firstName + " " + lastName))},
			{Name: aws.String("phone_number"), Value: aws.String(phone)},
			{Name: aws.String("email_verified"), Value: aws.String("true")},
		},
	})
	if err != nil {
		var exists *cognitoTypes.UsernameExistsException
		if errors.As(err, &exists) {
			return "", cognito.ErrUserAlreadyExists
		}
		return "", fmt.Errorf("admin create user: %w", err)
	}

	_, err = s.cognitoClient.Svc().AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(s.cognitoClient.PoolID()),
		Username:   aws.String(email),
		Password:   aws.String(password),
		Permanent:  true,
	})
	if err != nil {
		return "", fmt.Errorf("set permanent password: %w", err)
	}

	userSub := findAttributeValue(output.User.Attributes, "sub")
	if userSub == "" {
		return "", errors.New("missing cognito user sub")
	}

	return userSub, nil
}

func validateInvite(invite repository.Invite) error {
	if invite.IsUsed || invite.UsedAt != nil {
		return ErrInviteAlreadyUsed
	}
	if time.Now().UTC().After(invite.ExpiresAt.UTC()) {
		return ErrInviteExpired
	}

	return nil
}

func resolveRole(role *string) string {
	if role == nil || strings.TrimSpace(*role) == "" {
		return defaultInviteRole
	}

	return strings.TrimSpace(*role)
}

func trimOptional(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	return strings.ContainsAny(password, "!@#$%^&*")
}

func normalizeInviteCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func generateInviteCode() (string, error) {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const digits = "0123456789"

	var firstPart strings.Builder
	for range 4 {
		index, err := randomIndex(len(letters))
		if err != nil {
			return "", err
		}
		firstPart.WriteByte(letters[index])
	}

	var secondPart strings.Builder
	for range 4 {
		index, err := randomIndex(len(digits))
		if err != nil {
			return "", err
		}
		secondPart.WriteByte(digits[index])
	}

	return firstPart.String() + "-" + secondPart.String(), nil
}

func randomIndex(max int) (int, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}

	return int(n.Int64()), nil
}

func findAttributeValue(attrs []cognitoTypes.AttributeType, key string) string {
	for _, attr := range attrs {
		if aws.ToString(attr.Name) == key {
			return aws.ToString(attr.Value)
		}
	}

	return ""
}
