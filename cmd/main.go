// @title           HRMS API
// @version         1.0
// @description     Human Resource Management System API
// @host            localhost:8080
// @BasePath        /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.

package main

import (
	_ "hrms/docs"
	attendanceRepository "hrms/internal/feature/attendance/repository"
	attendanceService "hrms/internal/feature/attendance/service"
	attendanceHandler "hrms/internal/feature/attendance/transport/http"
	authRepository "hrms/internal/feature/auth/repository"
	authService "hrms/internal/feature/auth/service"
	authHandler "hrms/internal/feature/auth/transport/http"
	calendarRepository "hrms/internal/feature/calendar/repository"
	calendarService "hrms/internal/feature/calendar/service"
	calendarHandler "hrms/internal/feature/calendar/transport/http"
	cityRepository "hrms/internal/feature/city/repository"
	cityService "hrms/internal/feature/city/service"
	cityHandler "hrms/internal/feature/city/transport/http"
	consentRepository "hrms/internal/feature/consent/repository"
	consentService "hrms/internal/feature/consent/service"
	employeeRepository "hrms/internal/feature/employee/repository"
	employeeService "hrms/internal/feature/employee/service"
	employeeHandler "hrms/internal/feature/employee/transport/http"
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
	"hrms/internal/infrastructure/middleware"
	"hrms/internal/infrastructure/storage/postgres"
	"hrms/pkg/log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

	notificationRepo := notificationRepository.NewNotificationRepository(postgres.DB)
	notificationSvc := notificationService.NewService(notificationRepo)

	calendarRepo := calendarRepository.NewRepository(postgres.DB)
	calendarSvc := calendarService.NewCalendarService(calendarRepo, notificationSvc)
	calendarHTTPHandler := calendarHandler.NewCalendarHandler(calendarSvc, cognitoSvc)

	attendanceRepo := attendanceRepository.NewRepository(postgres.DB, calendarRepo)
	attendanceSvc := attendanceService.NewAttendanceService(attendanceRepo, notificationSvc)
	attendanceHTTPHandler := attendanceHandler.NewAttendanceHandler(attendanceSvc, cognitoSvc)

	cityRepo := cityRepository.NewRepository(postgres.DB)
	citySvc := cityService.NewCityService(cityRepo)
	cityHTTPHandler := cityHandler.NewCityHandler(citySvc)

	orgRepo := oganizationRepository.NewOrganizationRepository(postgres.DB)
	consentRepo := consentRepository.NewRepository(postgres.DB)
	inviteRepo := inviteRepository.NewRepository(postgres.DB)
	authRepo := authRepository.NewAuthRepository(postgres.DB)
	userRepo := userRepository.NewRepository(postgres.DB)
	eventsRepo := eventsRepository.NewRepository(postgres.DB)
	employeeRepo := employeeRepository.NewRepository(postgres.DB)

	authSvc := authService.NewAuthService(cognitoSvc, authRepo)
	consentSvc := consentService.NewConsentService(consentRepo)
	userSvc := userService.NewService(cognitoSvc, userRepo)
	eventsSvc := eventsService.NewService(eventsRepo)
	orgSvc := organizationService.NewSignUpService(orgRepo, consentRepo, notificationSvc, cognitoSvc, emailSvc)
	employeeSvc := employeeService.NewEmployeeService(employeeRepo)
	inviteSvc, err := inviteService.NewService(inviteRepo, notificationSvc, cfg, cognitoClient)
	if err != nil {
		logger.Fatal("Failed to initialize Invite service")
	}

	newAuthHTTPHandler := authHandler.NewAuthHandler(authSvc)
	employeeHTTPHandler := employeeHandler.NewEmployeeHandler(employeeSvc, cognitoSvc)
	handler := organizationHandler.NewOrganizationHandler(orgSvc, consentSvc)
	inviteHTTPHandler := inviteHandler.NewHandler(inviteSvc)
	userHTTPHandler := userHandler.NewHandler(userSvc)
	eventsHTTPHandler := eventsHandler.NewHandler(eventsSvc)
	notificationHTTPHandler := notificationHandler.NewHandler(notificationSvc)

	authMw, err := middleware.AuthMiddleware(cfg.AWS.Region, cfg.Cognito.UserPoolID, employeeRepo)
	if err != nil {
		logger.Fatal("Failed to initialize auth middleware")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	v1 := router.Group("/v1")

	// Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	v1.POST("/auth/login", newAuthHTTPHandler.Login)
	v1.POST("/auth/refresh", newAuthHTTPHandler.RefreshTokens)
	v1.GET("/profile/me", userHTTPHandler.Me)
	v1.POST("/auth/forgot-password", newAuthHTTPHandler.ForgotPassword)
	v1.POST("/auth/reset-password", newAuthHTTPHandler.ResetPassword)

	// Organizations
	v1.POST("/organizations", handler.CreateOrganization)
	v1.POST("/organizations/verify-otp", handler.VerifyOTP)
	v1.POST("/invites/generate", inviteHTTPHandler.GenerateInvite)
	v1.POST("/invites/verify", inviteHTTPHandler.VerifyInvite)
	v1.POST("/invites/complete-registration", inviteHTTPHandler.CompleteRegistration)
	v1.GET("/legal/documents", handler.GetDocuments)
	v1.GET("/cities", cityHTTPHandler.ListCities)

	// Protected routes
	protected := v1.Group("/", authMw)

	// Auth (protected)
	protected.POST("/auth/logout", newAuthHTTPHandler.Logout)

	// Events
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
	protected.POST("/organizations/consents", handler.SubmitConsents)
	protected.GET("/organizations/consents/validate", handler.ValidateConsents)

	// Departments
	protected.POST("/organizations/departments", handler.CreateDepartment)
	protected.GET("/organizations/departments", handler.ListDepartments)
	protected.DELETE("/organizations/departments/:id", handler.DeleteDepartment)

	// Positions
	protected.POST("/organizations/positions", handler.CreatePosition)
	protected.GET("/organizations/positions", handler.ListPositions)
	protected.DELETE("/organizations/positions/:id", handler.DeletePosition)

	// Employees
	protected.POST("/employees", employeeHTTPHandler.CreateEmployee)
	protected.GET("/employees", employeeHTTPHandler.ListEmployees)
	protected.GET("/departments", employeeHTTPHandler.ListDepartments)
	protected.GET("/positions", employeeHTTPHandler.ListPositions)
	protected.GET("/reference/departments", employeeHTTPHandler.ListDepartments)
	protected.GET("/reference/positions", employeeHTTPHandler.ListPositions)
	protected.GET("/references/departments", employeeHTTPHandler.ListDepartments)
	protected.GET("/references/positions", employeeHTTPHandler.ListPositions)
	protected.GET("/lookup/departments", employeeHTTPHandler.ListDepartments)
	protected.GET("/lookup/positions", employeeHTTPHandler.ListPositions)
	protected.GET("/employees/:id", employeeHTTPHandler.GetEmployee)
	protected.PATCH("/employees/:id/role", employeeHTTPHandler.UpdateRole)
	protected.PATCH("/employees/:id/salary", employeeHTTPHandler.UpdateSalary)
	protected.PATCH("/employees/:id/status", employeeHTTPHandler.UpdateStatus)
	protected.PATCH("/employees/:id/department", employeeHTTPHandler.UpdateDepartment)
	protected.PATCH("/employees/:id/position", employeeHTTPHandler.UpdatePosition)
	protected.DELETE("/employees/:id", employeeHTTPHandler.DeleteEmployee)

	// Attendance — work schedules
	protected.POST("/attendance/work-schedules", attendanceHTTPHandler.SetWorkSchedule)
	protected.GET("/attendance/work-schedules/:employee_id", attendanceHTTPHandler.GetWorkSchedule)

	// Attendance — СКУД events
	protected.POST("/attendance/skud-events", attendanceHTTPHandler.CreateSkudEvent)

	// Attendance — leave requests
	protected.POST("/attendance/leave-requests", attendanceHTTPHandler.CreateLeaveRequest)
	protected.GET("/attendance/leave-requests", attendanceHTTPHandler.ListLeaveRequests)
	protected.GET("/attendance/leave-requests/:id", attendanceHTTPHandler.GetLeaveRequest)
	protected.PATCH("/attendance/leave-requests/:id/review", attendanceHTTPHandler.ReviewLeaveRequest)

	// Attendance — records
	protected.GET("/attendance", attendanceHTTPHandler.ListAttendance)
	protected.GET("/attendance/employees/:employee_id", attendanceHTTPHandler.ListEmployeeAttendance)

	// Calendar
	protected.POST("/calendar", calendarHTTPHandler.AddDay)
	protected.DELETE("/calendar/:id", calendarHTTPHandler.DeleteDay)
	protected.GET("/calendar/summary", calendarHTTPHandler.GetMonthSummary)

	port := ":8080"
	logger.Info("Starting HTTP server on port " + port)
	if err := router.Run(port); err != nil {
		logger.Fatal("Server failed", log.Error(err))
	}
}
