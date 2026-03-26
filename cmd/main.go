package main

import (
	authRepository "hrms/internal/feature/auth/repository"
	authService "hrms/internal/feature/auth/service"
	authHandler "hrms/internal/feature/auth/transport/http"
	consentRepository "hrms/internal/feature/consent/repository"
	consentService "hrms/internal/feature/consent/service"
	employeeRepository "hrms/internal/feature/employee/repository"
	employeeService "hrms/internal/feature/employee/service"
	employeeHandler "hrms/internal/feature/employee/transport/http"
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
	authRepo := authRepository.NewAuthRepository(postgres.DB)
	employeeRepo := employeeRepository.NewRepository(postgres.DB)

	// TODO: Initialize services for all modules.
	authSvc := authService.NewAuthService(cognitoSvc, authRepo)
	consentSvc := consentService.NewConsentService(consentRepo)
	orgSvc := organizationService.NewSignUpService(orgRepo, consentRepo, cognitoSvc, emailSvc)
	employeeSvc := employeeService.NewEmployeeService(employeeRepo)
	inviteSvc, err := inviteService.NewInviteService(inviteRepo, cfg, cognitoClient)
	if err != nil {
		logger.Fatal("Failed to initialize Invite service")
	}

	// TODO: Initialize handlers for all modules
	newAuthHTTPHandler := authHandler.NewAuthHandler(authSvc)
	employeeHTTPHandler := employeeHandler.NewEmployeeHandler(employeeSvc, cognitoSvc)
	handler := organizationHandler.NewOrganizationHandler(orgSvc, consentSvc)
	inviteHTTPHandler := inviteHandler.NewHandler(inviteSvc)

	router := gin.Default()

	v1 := router.Group("/v1")

	// Auth
	v1.POST("/auth/login", newAuthHTTPHandler.Login)
	v1.POST("/auth/refresh", newAuthHTTPHandler.RefreshTokens)
	// Organizations
	v1.POST("/organizations", handler.CreateOrganization)
	v1.POST("/organizations/verify-otp", handler.VerifyOTP)
	v1.POST("/invites/generate", inviteHTTPHandler.GenerateInvite)
	v1.POST("/invites/verify", inviteHTTPHandler.VerifyInvite)
	v1.POST("/invites/complete-registration", inviteHTTPHandler.CompleteRegistration)

	// Consents
	v1.POST("/organizations/consents", handler.SubmitConsents)
	v1.GET("/organizations/consents/validate", handler.ValidateConsents)
	v1.GET("/legal/documents", handler.GetDocuments)

	// Employees
	v1.POST("/employees", employeeHTTPHandler.CreateEmployee)
	v1.GET("/employees", employeeHTTPHandler.ListEmployees)
	v1.GET("/employees/:id", employeeHTTPHandler.GetEmployee)
	v1.PATCH("/employees/:id/role", employeeHTTPHandler.UpdateRole)
	v1.PATCH("/employees/:id/salary", employeeHTTPHandler.UpdateSalary)
	v1.PATCH("/employees/:id/status", employeeHTTPHandler.UpdateStatus)
	v1.PATCH("/employees/:id/department", employeeHTTPHandler.UpdateDepartment)
	v1.PATCH("/employees/:id/position", employeeHTTPHandler.UpdatePosition)
	v1.DELETE("/employees/:id", employeeHTTPHandler.DeleteEmployee)

	port := ":8080"
	logger.Info("Starting HTTP server on port " + port)
	if err := router.Run(port); err != nil {
		logger.Fatal("Server failed", log.Error(err))
	}
}
