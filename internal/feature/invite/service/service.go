package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/mail"
	"regexp"
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

func NewInviteService(repo *repository.Repository, cfg *config.Config, cognitoClient *cognito.Client) (*Service, error) {
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
	email := strings.TrimSpace(req.Email)
	firstName := strings.TrimSpace(req.FirstName)
	lastName := strings.TrimSpace(req.LastName)

	log.Printf("[Invite Generate] Starting invite generation for org=%q email=%q", orgIDRaw, email)

	if orgIDRaw == "" {
		log.Printf("[Invite Generate] Validation failed: organization id is required")
		return nil, ErrOrganizationIDRequired
	}
	if firstName == "" {
		log.Printf("[Invite Generate] Validation failed: first name is required")
		return nil, ErrFirstNameRequired
	}
	if lastName == "" {
		log.Printf("[Invite Generate] Validation failed: last name is required")
		return nil, ErrLastNameRequired
	}
	if email == "" {
		log.Printf("[Invite Generate] Validation failed: email is required")
		return nil, ErrEmailRequired
	}
	if _, err := mail.ParseAddress(email); err != nil {
		log.Printf("[Invite Generate] Validation failed: invalid email format for %q", email)
		return nil, ErrInvalidEmail
	}

	orgID, err := uuid.Parse(orgIDRaw)
	if err != nil {
		log.Printf("[Invite Generate] Validation failed: invalid organization id %q", orgIDRaw)
		return nil, ErrInvalidOrganizationID
	}

	log.Printf("[Invite Generate] Looking up organization name for org=%s", orgID)
	organizationName, err := s.repo.GetOrganizationNameByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Invite Generate] Organization not found: org=%s", orgID)
			return nil, ErrOrganizationNotFound
		}
		log.Printf("[Invite Generate] Failed to load organization %s: %v", orgID, err)
		return nil, fmt.Errorf("get organization by id: %w", err)
	}
	log.Printf("[Invite Generate] Organization resolved: org=%s name=%q", orgID, organizationName)

	inviteID := uuid.New()
	expiresAt := time.Now().UTC().Add(inviteTTL)

	var code string
	for attempt := 1; attempt <= maxInviteAttempts; attempt++ {
		code, err = generateInviteCode()
		if err != nil {
			log.Printf("[Invite Generate] Failed to generate invite code on attempt=%d: %v", attempt, err)
			return nil, fmt.Errorf("generate invite code: %w", err)
		}

		log.Printf("[Invite Generate] Attempt=%d generated code=%s for email=%q", attempt, code, email)

		insertErr := s.repo.CreateInvite(ctx, repository.CreateInviteParams{
			ID:        inviteID,
			OrgID:     orgID,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Code:      code,
			Role:      resolveRole(req.Role),
			Position:  trimOptional(req.Position),
			ExpiresAt: expiresAt,
		})
		if insertErr == nil {
			log.Printf("[Invite Generate] Invite persisted: inviteID=%s code=%s email=%q expiresAt=%s", inviteID, code, email, expiresAt.Format(time.RFC3339))
			err = nil
			break
		}
		if !repository.IsUniqueViolation(insertErr) {
			log.Printf("[Invite Generate] Failed to persist invite for email=%q: %v", email, insertErr)
			return nil, fmt.Errorf("create invite: %w", insertErr)
		}
		log.Printf("[Invite Generate] Code collision on attempt=%d for code=%s", attempt, code)
		err = insertErr
	}
	if err != nil {
		log.Printf("[Invite Generate] Exhausted invite generation attempts for email=%q", email)
		return nil, ErrGenerateInvite
	}

	log.Printf("[Invite Generate] Sending invite email via SMTP to email=%q code=%s", email, code)
	if err := s.mailer.SendInvite(
		ctx,
		email,
		firstName,
		organizationName,
		code,
		defaultPortalURL,
	); err != nil {
		log.Printf("[Invite Generate] Invite email send failed for inviteID=%s email=%q: %v", inviteID, email, err)
		_ = s.repo.DeleteInviteByID(ctx, inviteID)
		log.Printf("[Invite Generate] Invite rolled back after email failure: inviteID=%s", inviteID)
		return nil, fmt.Errorf("send invite email: %w", err)
	}
	log.Printf("[Invite Generate] Invite email sent successfully: inviteID=%s email=%q", inviteID, email)

	return &GenerateInviteResponse{
		InviteID:         inviteID.String(),
		OrganizationID:   orgID.String(),
		OrganizationName: organizationName,
		Email:            email,
		Code:             code,
		ExpiresAt:        expiresAt,
	}, nil
}

func (s *Service) VerifyInvite(ctx context.Context, req VerifyInviteRequest) (*VerifyInviteResponse, error) {
	code := normalizeInviteCode(req.Code)
	log.Printf("[Invite Verify] Verifying invite code=%q", code)
	if code == "" {
		log.Printf("[Invite Verify] Validation failed: invite code is required")
		return nil, ErrInviteCodeRequired
	}

	invite, err := s.repo.GetInviteByCode(ctx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Invite Verify] Invite not found for code=%q", code)
			return nil, ErrInviteNotFound
		}
		log.Printf("[Invite Verify] Failed to load invite for code=%q: %v", code, err)
		return nil, fmt.Errorf("get invite by code: %w", err)
	}

	if err := validateInvite(invite); err != nil {
		log.Printf("[Invite Verify] Invite validation failed for code=%q email=%q: %v", code, invite.Email, err)
		return nil, err
	}

	log.Printf("[Invite Verify] Invite is valid: code=%q email=%q org=%q", code, invite.Email, invite.OrganizationName)

	return &VerifyInviteResponse{
		OrganizationID:   invite.OrgID,
		OrganizationName: invite.OrganizationName,
		FirstName:        invite.FirstName,
		LastName:         invite.LastName,
		FullName:         strings.TrimSpace(invite.FirstName + " " + invite.LastName),
		Email:            invite.Email,
		Role:             invite.Role,
		Position:         invite.Position,
		ExpiresAt:        invite.ExpiresAt,
		Message:          fmt.Sprintf("You have been invited to join %s", invite.OrganizationName),
	}, nil
}

func (s *Service) CompleteRegistration(ctx context.Context, req CompleteRegistrationRequest) (*CompleteRegistrationResponse, error) {
	code := normalizeInviteCode(req.Code)
	log.Printf("[Invite CompleteRegistration] Starting registration for code=%q", code)
	if code == "" {
		log.Printf("[Invite CompleteRegistration] Validation failed: invite code is required")
		return nil, ErrInviteCodeRequired
	}
	if !isValidPassword(req.Password) {
		log.Printf("[Invite CompleteRegistration] Validation failed for code=%q: password does not meet policy", code)
		return nil, ErrPasswordRequired
	}
	if strings.TrimSpace(req.PhoneNumber) == "" {
		log.Printf("[Invite CompleteRegistration] Validation failed for code=%q: phone number is required", code)
		return nil, ErrPhoneNumberRequired
	}
	if !isValidPhoneNumber(req.PhoneNumber) {
		log.Printf("[Invite CompleteRegistration] Validation failed for code=%q: invalid phone number format %q", code, req.PhoneNumber)
		return nil, ErrInvalidPhoneNumber
	}

	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		log.Printf("[Invite CompleteRegistration] Failed to begin transaction for code=%q: %v", code, err)
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	invite, err := s.repo.GetInviteByCodeTx(ctx, tx, code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Invite CompleteRegistration] Invite not found for code=%q", code)
			return nil, ErrInviteNotFound
		}
		log.Printf("[Invite CompleteRegistration] Failed to lock invite for code=%q: %v", code, err)
		return nil, fmt.Errorf("lock invite by code: %w", err)
	}

	if err := validateInvite(invite); err != nil {
		log.Printf("[Invite CompleteRegistration] Invite validation failed for code=%q email=%q: %v", code, invite.Email, err)
		return nil, err
	}

	firstName := invite.FirstName
	lastName := invite.LastName
	phoneNumber := strings.TrimSpace(req.PhoneNumber)

	log.Printf("[Invite CompleteRegistration] Creating Cognito user for code=%q email=%q", code, invite.Email)

	userSub, err := s.createConfirmedUser(
		ctx,
		invite.Email,
		req.Password,
		firstName,
		lastName,
		phoneNumber,
	)
	if err != nil {
		if errors.Is(err, cognito.ErrUserAlreadyExists) {
			log.Printf("[Invite CompleteRegistration] Cognito user already exists for code=%q email=%q", code, invite.Email)
			return nil, ErrUserAlreadyExists
		}
		log.Printf("[Invite CompleteRegistration] Failed to create Cognito user for code=%q email=%q: %v", code, invite.Email, err)
		return nil, fmt.Errorf("create cognito user: %w", err)
	}
	log.Printf("[Invite CompleteRegistration] Cognito user created: code=%q email=%q sub=%s", code, invite.Email, userSub)

	userID, err := uuid.Parse(userSub)
	if err != nil {
		log.Printf("[Invite CompleteRegistration] Invalid Cognito sub for code=%q email=%q: %v", code, invite.Email, err)
		s.rollbackCreatedCognitoUser(ctx, invite.Email, "invalid cognito sub")
		return nil, fmt.Errorf("parse cognito user id: %w", err)
	}
	orgID, err := uuid.Parse(invite.OrgID)
	if err != nil {
		log.Printf("[Invite CompleteRegistration] Invalid organization id on invite code=%q: %v", code, err)
		s.rollbackCreatedCognitoUser(ctx, invite.Email, "invalid organization id")
		return nil, fmt.Errorf("parse org id: %w", err)
	}
	inviteID, err := uuid.Parse(invite.ID)
	if err != nil {
		log.Printf("[Invite CompleteRegistration] Invalid invite id for code=%q: %v", code, err)
		s.rollbackCreatedCognitoUser(ctx, invite.Email, "invalid invite id")
		return nil, fmt.Errorf("parse invite id: %w", err)
	}

	log.Printf("[Invite CompleteRegistration] Inserting application user for code=%q email=%q", code, invite.Email)
	if err := s.repo.InsertUserTx(ctx, tx, repository.CreateInvitedUserParams{
		ID:          userID,
		OrgID:       orgID,
		Email:       invite.Email,
		Role:        invite.Role,
		FirstName:   firstName,
		LastName:    lastName,
		PhoneNumber: phoneNumber,
	}); err != nil {
		log.Printf("[Invite CompleteRegistration] Failed to insert application user for code=%q email=%q: %v", code, invite.Email, err)

		s.rollbackCreatedCognitoUser(ctx, invite.Email, "application user insert failed")

		if repository.IsUniqueViolation(err) {
			switch repository.UniqueConstraintName(err) {
			case "users_phone_number_key":
				return nil, ErrPhoneNumberExists
			case "users_email_key":
				return nil, ErrEmailAlreadyExists
			}
		}

		return nil, fmt.Errorf("insert invited user: %w", err)
	}

	log.Printf("[Invite CompleteRegistration] Marking invite as used: inviteID=%s code=%q", inviteID, code)
	if err := s.repo.MarkInviteUsedTx(ctx, tx, inviteID, time.Now().UTC()); err != nil {
		log.Printf("[Invite CompleteRegistration] Failed to mark invite as used for code=%q: %v", code, err)
		s.rollbackCreatedCognitoUser(ctx, invite.Email, "mark invite used failed")
		return nil, fmt.Errorf("mark invite used: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[Invite CompleteRegistration] Transaction commit failed for code=%q email=%q: %v", code, invite.Email, err)
		s.rollbackCreatedCognitoUser(ctx, invite.Email, "transaction commit failed")
		return nil, fmt.Errorf("commit registration: %w", err)
	}

	log.Printf("[Invite CompleteRegistration] Registration completed successfully: code=%q email=%q userID=%s", code, invite.Email, userID)

	return &CompleteRegistrationResponse{
		UserID:         userID.String(),
		OrganizationID: invite.OrgID,
		Role:           invite.Role,
	}, nil
}

func (s *Service) createConfirmedUser(ctx context.Context, email, password, firstName, lastName, phone string) (string, error) {
	log.Printf("[Invite Cognito] AdminCreateUser started for email=%q", email)
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
			log.Printf("[Invite Cognito] AdminCreateUser reported existing user for email=%q", email)
			return "", cognito.ErrUserAlreadyExists
		}
		var invalidParameterErr *cognitoTypes.InvalidParameterException
		if errors.As(err, &invalidParameterErr) {
			message := strings.TrimSpace(aws.ToString(invalidParameterErr.Message))
			if strings.Contains(strings.ToLower(message), "phone number") {
				log.Printf("[Invite Cognito] Invalid phone number rejected by Cognito for email=%q", email)
				return "", ErrInvalidPhoneNumber
			}
		}
		log.Printf("[Invite Cognito] AdminCreateUser failed for email=%q: %v", email, err)
		return "", fmt.Errorf("admin create user: %w", err)
	}
	log.Printf("[Invite Cognito] AdminCreateUser succeeded for email=%q", email)

	log.Printf("[Invite Cognito] AdminSetUserPassword started for email=%q", email)
	_, err = s.cognitoClient.Svc().AdminSetUserPassword(ctx, &cognitoidentityprovider.AdminSetUserPasswordInput{
		UserPoolId: aws.String(s.cognitoClient.PoolID()),
		Username:   aws.String(email),
		Password:   aws.String(password),
		Permanent:  true,
	})
	if err != nil {
		log.Printf("[Invite Cognito] AdminSetUserPassword failed for email=%q: %v", email, err)

		var invalidPasswordErr *cognitoTypes.InvalidPasswordException
		if errors.As(err, &invalidPasswordErr) {
			log.Printf("[Invite Cognito] Rolling back Cognito user after invalid password for email=%q", email)
			_, deleteErr := s.cognitoClient.Svc().AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
				UserPoolId: aws.String(s.cognitoClient.PoolID()),
				Username:   aws.String(email),
			})
			if deleteErr != nil {
				log.Printf("[Invite Cognito] Failed to roll back Cognito user for email=%q: %v", email, deleteErr)
			} else {
				log.Printf("[Invite Cognito] Rolled back Cognito user for email=%q after invalid password", email)
			}

			message := strings.TrimSpace(aws.ToString(invalidPasswordErr.Message))
			if message == "" {
				message = "password does not conform to policy"
			}

			return "", &PasswordPolicyError{
				Message: message,
			}
		}

		return "", fmt.Errorf("set permanent password: %w", err)
	}
	log.Printf("[Invite Cognito] AdminSetUserPassword succeeded for email=%q", email)

	userSub := findAttributeValue(output.User.Attributes, "sub")
	if userSub == "" {
		log.Printf("[Invite Cognito] Missing sub attribute for email=%q", email)
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

func isValidPhoneNumber(phone string) bool {
	trimmed := strings.TrimSpace(phone)
	matched, _ := regexp.MatchString(`^\+[1-9]\d{7,14}$`, trimmed)
	return matched
}

func (s *Service) rollbackCreatedCognitoUser(ctx context.Context, email, reason string) {
	log.Printf("[Invite Cognito] Rolling back created Cognito user for email=%q reason=%q", email, reason)

	_, err := s.cognitoClient.Svc().AdminDeleteUser(ctx, &cognitoidentityprovider.AdminDeleteUserInput{
		UserPoolId: aws.String(s.cognitoClient.PoolID()),
		Username:   aws.String(email),
	})
	if err != nil {
		log.Printf("[Invite Cognito] Failed to roll back Cognito user for email=%q: %v", email, err)
		return
	}

	log.Printf("[Invite Cognito] Rolled back Cognito user for email=%q", email)
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
