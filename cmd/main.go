package main

import (
	consentRepository "hrms/internal/feature/consent/repository"
	consentService "hrms/internal/feature/consent/service"
	inviteRepository "hrms/internal/feature/invite/repository"
	inviteService "hrms/internal/feature/invite/service"
	inviteHandler "hrms/internal/feature/invite/transport/http"
	oganizationRepository "hrms/internal/feature/organization/repository"
	organizationService "hrms/internal/feature/organization/service"
	organizationHandler "hrms/internal/feature/organization/transport/http"
	"hrms/internal/infrastructure/app/cognito"
	"hrms/internal/infrastructure/config"
	"hrms/internal/infrastructure/email"
	"hrms/internal/infrastructure/storage/postgres"
	"hrms/pkg/log"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.ParseConfig("")
	logger := log.NewLog(cfg.LogLevel)
	postgres.InitDB(cfg)
	cognitoClient, err := cognito.New(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize Cognito client")
	}
	cognitoSvc := cognito.NewService(cognitoClient)
	emailSvc, err := email.NewService(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize Email client")
	}

	// TODO: Initialize repositories for all modules.
	orgRepo := oganizationRepository.NewOrganizationRepository(postgres.DB)
	consentRepo := consentRepository.NewRepository(postgres.DB)
	inviteRepo := inviteRepository.NewRepository(postgres.DB)

	// TODO: Initialize services for all modules.
	consentSvc := consentService.NewService(consentRepo)
	orgSvc := organizationService.NewSignUpService(orgRepo, consentRepo, cognitoSvc, emailSvc)
	inviteSvc, err := inviteService.NewService(inviteRepo, cfg, cognitoClient)
	if err != nil {
		logger.Fatal("Failed to initialize Invite service")
	}

	handler := organizationHandler.NewOrganizationHandler(orgSvc, consentSvc)
	inviteHTTPHandler := inviteHandler.NewHandler(inviteSvc)

	router := gin.Default()

	v1 := router.Group("/v1")

	v1.POST("/organizations", handler.CreateOrganization)
	v1.POST("/organizations/verify-otp", handler.VerifyOTP)
	v1.POST("/invites/generate", inviteHTTPHandler.GenerateInvite)
	v1.POST("/invites/verify", inviteHTTPHandler.VerifyInvite)
	v1.POST("/invites/complete-registration", inviteHTTPHandler.CompleteRegistration)

	v1.POST("/organizations/consents", handler.SubmitConsents)
	v1.GET("/organizations/consents/validate", handler.ValidateConsents)
	v1.GET("/legal/documents", handler.GetDocuments)

	port := ":8080"
	logger.Info("Starting HTTP server on port " + port)
	if err := router.Run(port); err != nil {
		logger.Fatal("Server failed", log.Error(err))
	}
}
