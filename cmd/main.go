package main

import (
	authRepository "hrms/internal/feature/auth/repository"
	authService "hrms/internal/feature/auth/service"
	authHandler "hrms/internal/feature/auth/transport/http"
	consentRepository "hrms/internal/feature/consent/repository"
	consentService "hrms/internal/feature/consent/service"
	eventsRepository "hrms/internal/feature/events/repository"
	eventsService "hrms/internal/feature/events/service"
	eventsHandler "hrms/internal/feature/events/transport/http"
	inviteRepository "hrms/internal/feature/invite/repository"
	inviteService "hrms/internal/feature/invite/service"
	inviteHandler "hrms/internal/feature/invite/transport/http"
	notificationRepository "hrms/internal/feature/notification/repository/postgres"
	notificationService "hrms/internal/feature/notification/service"
	notificationHandler "hrms/internal/feature/notification/transport/http"
	oganizationRepository "hrms/internal/feature/organization/repository"
	organizationService "hrms/internal/feature/organization/service"
	organizationHandler "hrms/internal/feature/organization/transport/http"
	userRepository "hrms/internal/feature/user/repository"
	userService "hrms/internal/feature/user/service"
	userHandler "hrms/internal/feature/user/transport/http"
	"hrms/internal/infrastructure/app/cognito"
	"hrms/internal/infrastructure/config"
	"hrms/internal/infrastructure/email"
	"hrms/internal/infrastructure/storage/postgres"
	"hrms/pkg/log"

	"github.com/gin-contrib/cors"
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
	userRepo := userRepository.NewRepository(postgres.DB)
	eventsRepo := eventsRepository.NewRepository(postgres.DB)
	notificationRepo := notificationRepository.NewNotificationRepository(postgres.DB)

	// TODO: Initialize services for all modules.
	authSvc := authService.NewAuthService(cognitoSvc, authRepo)
	consentSvc := consentService.NewService(consentRepo)
	userSvc := userService.NewService(cognitoSvc, userRepo)
	eventsSvc := eventsService.NewService(eventsRepo)
	notificationSvc := notificationService.NewService(notificationRepo)
	orgSvc := organizationService.NewSignUpService(orgRepo, consentRepo, notificationSvc, cognitoSvc, emailSvc)
	inviteSvc, err := inviteService.NewService(inviteRepo, notificationSvc, cfg, cognitoClient)
	if err != nil {
		logger.Fatal("Failed to initialize Invite service")
	}

	newAuthHandler := authHandler.NewAuthHandler(authSvc)
	handler := organizationHandler.NewOrganizationHandler(orgSvc, consentSvc)
	inviteHTTPHandler := inviteHandler.NewHandler(inviteSvc)
	userHTTPHandler := userHandler.NewHandler(userSvc)
	eventsHTTPHandler := eventsHandler.NewHandler(eventsSvc)
	notificationHTTPHandler := notificationHandler.NewHandler(notificationSvc)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := router.Group("/v1")

	// Auth
	v1.POST("/auth/login", newAuthHandler.Login)
	v1.POST("/auth/refresh", newAuthHandler.RefreshTokens)
	v1.GET("/profile/me", userHTTPHandler.Me)
	// Organizations
	v1.POST("/organizations", handler.CreateOrganization)
	v1.POST("/organizations/verify-otp", handler.VerifyOTP)
	v1.POST("/invites/generate", inviteHTTPHandler.GenerateInvite)
	v1.POST("/invites/verify", inviteHTTPHandler.VerifyInvite)
	v1.POST("/invites/complete-registration", inviteHTTPHandler.CompleteRegistration)

	events := v1.Group("/events")
	events.Use(cognito.AuthMiddleware(cognitoSvc, userRepo))
	events.GET("/upcoming", eventsHTTPHandler.Upcoming)
	events.GET("/my", eventsHTTPHandler.My)
	events.POST("", eventsHTTPHandler.Create)
	events.PATCH("/:id", eventsHTTPHandler.Update)
	events.DELETE("/:id", eventsHTTPHandler.Delete)

	// Notifications
	notifications := v1.Group("/notifications")
	notifications.Use(cognito.AuthMiddleware(cognitoSvc, userRepo))
	notifications.GET("", notificationHTTPHandler.ListNotifications)
	notifications.PATCH("/:id/read", notificationHTTPHandler.MarkAsRead)
	notifications.PATCH("/read-all", notificationHTTPHandler.MarkAllAsRead)

	// Consents
	v1.POST("/organizations/consents", handler.SubmitConsents)
	v1.GET("/organizations/consents/validate", handler.ValidateConsents)
	v1.GET("/legal/documents", handler.GetDocuments)

	port := ":8080"
	logger.Info("Starting HTTP server on port " + port)
	if err := router.Run(port); err != nil {
		logger.Fatal("Server failed", log.Error(err))
	}
}
