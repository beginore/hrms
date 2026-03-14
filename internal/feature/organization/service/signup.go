package service

import (
	"context"
	"errors"
	"fmt"
	consentRepository "hrms/internal/feature/consent/repository"
	"hrms/internal/feature/organization/repository"
	"hrms/internal/feature/organization/repository/postgres"
	"hrms/internal/infrastructure/app/cognito"
	"hrms/internal/infrastructure/email"
	"log"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmailAlreadyExists   = errors.New("email already exists")
	ErrVATAlreadyExists     = errors.New("VAT already registered")
	ErrPoliciesNotAccepted  = errors.New("must accept privacy policy and terms")
	ErrCognitoFailed        = errors.New("failed to create user in Cognito")
	ErrDatabaseInsertFailed = errors.New("failed to save data")
	ErrDatabaseQueryFailed  = errors.New("failed to query database")
	ErrUserAlreadyExists    = errors.New("user already exists in Cognito")
)

type SignUpService struct {
	repo        repository.OrganizationRepository
	consentRepo consentRepository.ConsentRepository
	cognitoSvc  *cognito.Service
	emailSvc    *email.Service
}

func NewSignUpService(
	repo repository.OrganizationRepository,
	consentRepo consentRepository.ConsentRepository,
	cognitoSvc *cognito.Service,
	emailSvc *email.Service,
) *SignUpService {
	return &SignUpService{
		repo:        repo,
		consentRepo: consentRepo,
		cognitoSvc:  cognitoSvc,
		emailSvc:    emailSvc,
	}
}

func (s *SignUpService) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	log.Printf("[SignUp] Starting organization creation for email: %s", req.Email)

	if err := ValidateOrganization(ctx, req); err != nil {
		log.Printf("[SignUp] Validation failed: %v", err)
		return nil, err
	}

	log.Printf("[SignUp] Checking VAT uniqueness: %s", req.Vat)
	vatCount, err := s.repo.CheckVATUnique(ctx, req.Vat)
	if err != nil {
		log.Printf("[SignUp] VAT check failed: %v", err)
		return nil, fmt.Errorf("%w: vat check: %v", ErrDatabaseQueryFailed, err)
	}
	if vatCount > 0 {
		log.Printf("[SignUp] VAT already registered: %s", req.Vat)
		return nil, ErrVATAlreadyExists
	}
	if !req.PrivacyPolicyAccepted || !req.TermsAndConditionsAccepted {
		log.Printf("[SignUp] Policies not accepted for email: %s", req.Email)
		return nil, ErrPoliciesNotAccepted
	}

	log.Printf("[SignUp] Creating Cognito user: %s", req.Email)
	orgID, userID, err := s.createCognitoUser(ctx, req)
	if err != nil {
		log.Printf("[SignUp] Cognito user creation failed: %v", err)
		return nil, err
	}

	log.Printf("[SignUp] Persisting to DB - orgID: %s, userID: %s", orgID, userID)
	if err := s.persistToDatabase(ctx, req, orgID, userID); err != nil {
		log.Printf("[SignUp] DB persistence failed: %v", err)
		return nil, err
	}

	otp := s.emailSvc.GenerateOTP()
	if err := s.emailSvc.SendOTP(ctx, req.Email, otp); err != nil {
		log.Printf("[SignUp] OTP send failed (non-fatal): %v", err)
	}

	log.Printf("[SignUp] Organization created successfully: %s", orgID)
	return &CreateOrganizationResponse{OrganizationId: orgID.String()}, nil
}

func (s *SignUpService) createCognitoUser(ctx context.Context, req CreateOrganizationRequest) (uuid.UUID, uuid.UUID, error) {
	orgID := uuid.New()

	log.Printf("[Cognito] Signing up user: %s", req.Email)
	cognitoSub, err := s.cognitoSvc.SignUpUser(
		ctx, req.Email, req.Password, req.Firstname, req.Lastname, req.PhoneNumber,
	)
	if err != nil {
		if errors.Is(err, cognito.ErrUserAlreadyExists) {
			return uuid.Nil, uuid.Nil, ErrEmailAlreadyExists
		}
		return uuid.Nil, uuid.Nil, fmt.Errorf("%w: %v", ErrCognitoFailed, err)
	}

	return orgID, uuid.MustParse(cognitoSub), nil
}

func (s *SignUpService) VerifyOTP(ctx context.Context, req VerifyOTPRequest) error {
	log.Printf("[VerifyOTP] Verifying OTP for email: %s", req.Email)

	if err := s.cognitoSvc.ConfirmSignUp(ctx, req.Email, req.Code); err != nil {
		log.Printf("[VerifyOTP] Cognito ConfirmSignUp failed: %v", err)
		return err
	}

	log.Printf("[VerifyOTP] Updating verification status: %s", req.Email)
	if err := s.repo.UpdateUserVerificationStatus(ctx, req.Email); err != nil {
		log.Printf("[VerifyOTP] Failed to update verification status: %v", err)
		return fmt.Errorf("%w: update verification status: %v", ErrDatabaseInsertFailed, err)
	}

	log.Printf("[VerifyOTP] Verification complete: %s", req.Email)
	return nil
}

func (s *SignUpService) persistToDatabase(ctx context.Context, req CreateOrganizationRequest, orgID, userID uuid.UUID) error {
	log.Printf("[DB] Inserting organization: %s", orgID.String())
	if err := s.repo.CreateOrganization(ctx, postgres.InsertOrganizationParams{
		ID:          orgID,
		AdminID:     userID,
		Name:        req.OrganizationName,
		VatID:       req.Vat,
		Description: req.OrganizationDescription,
		Address:     req.StreetAddress,
		CityID:      req.CityId,
	}); err != nil {
		return fmt.Errorf("%w: org insert: %v", ErrDatabaseInsertFailed, err)
	}
	log.Printf("[DB] Inserting user: %s", userID)
	if err := s.repo.CreateUser(ctx, postgres.InsertUserParams{
		ID:          userID,
		OrgID:       orgID,
		Email:       req.Email,
		Role:        "SysAdmin",
		FirstName:   req.Firstname,
		LastName:    req.Lastname,
		PhoneNumber: req.PhoneNumber,
	}); err != nil {
		return fmt.Errorf("%w: user insert: %v", ErrDatabaseInsertFailed, err)
	}
	log.Printf("[DB] Fetching active documents")
	activeDocs, err := s.consentRepo.GetActiveDocuments(ctx)
	if err != nil {
		return fmt.Errorf("%w: fetch active documents: %v", ErrDatabaseQueryFailed, err)
	}

	log.Printf("[DB] Inserting consents for user: %s", userID)
	for _, doc := range activeDocs {
		if err := s.consentRepo.InsertConsentForOrg(ctx, consentRepository.InsertConsentForOrgParams{
			ID:           uuid.New(),
			UserID:       userID,
			OrgID:        orgID,
			DocumentType: doc.Type,
			Version:      doc.Version,
		}); err != nil {
			return fmt.Errorf("%w: consent insert for %s: %v", ErrDatabaseInsertFailed, doc.Type, err)
		}
	}

	log.Printf("[DB] All inserts successful")
	return nil
}
