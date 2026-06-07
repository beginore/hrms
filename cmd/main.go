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
	aiService "hrms/internal/feature/ai/service"
	aiHandler "hrms/internal/feature/ai/transport/http"
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
	payrollRepository "hrms/internal/feature/payroll/repository"
	payrollService "hrms/internal/feature/payroll/service"
	payrollHandler "hrms/internal/feature/payroll/transport/http"
	payslipRepository "hrms/internal/feature/payslip/repository"
	payslipService "hrms/internal/feature/payslip/service"
	payslipHandler "hrms/internal/feature/payslip/transport/http"
	reportsRepository "hrms/internal/feature/reports/repository"
	reportsService "hrms/internal/feature/reports/service"
	reportsHandler "hrms/internal/feature/reports/transport/http"
	taskRepository "hrms/internal/feature/task/repository"
	taskService "hrms/internal/feature/task/service"
	taskHandler "hrms/internal/feature/task/transport/http"
	userRepository "hrms/internal/feature/user/repository"
	userService "hrms/internal/feature/user/service"
	userHandler "hrms/internal/feature/user/transport/http"
	"hrms/internal/infrastructure/app/anthropic"
	"hrms/internal/infrastructure/app/cognito"
	"hrms/internal/infrastructure/config"
	"hrms/internal/infrastructure/email"
	"hrms/internal/infrastructure/middleware"
	"hrms/internal/infrastructure/storage/postgres"
	"hrms/pkg/log"
	"os"
	"time"

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

	calendarRepo := calendarRepository.NewRepository(postgres.DB)
	notificationRepo := notificationRepository.NewNotificationRepository(postgres.DB)
	notificationSvc := notificationService.NewService(notificationRepo)
	calendarSvc := calendarService.NewCalendarService(calendarRepo, notificationSvc)
	calendarHTTPHandler := calendarHandler.NewCalendarHandler(calendarSvc, cognitoSvc)

	attendanceRepo := attendanceRepository.NewRepository(postgres.DB, calendarRepo)
	attendanceSvc := attendanceService.NewAttendanceService(attendanceRepo, notificationSvc)
	attendanceHTTPHandler := attendanceHandler.NewAttendanceHandler(attendanceSvc)

	cityRepo := cityRepository.NewRepository(postgres.DB)
	citySvc := cityService.NewCityService(cityRepo)
	cityHTTPHandler := cityHandler.NewCityHandler(citySvc)

	orgRepo := oganizationRepository.NewOrganizationRepository(postgres.DB)
	consentRepo := consentRepository.NewRepository(postgres.DB)
	inviteRepo := inviteRepository.NewRepository(postgres.DB)
	authRepo := authRepository.NewAuthRepository(postgres.DB)
	employeeRepo := employeeRepository.NewRepository(postgres.DB)
	userRepo := userRepository.NewRepository(postgres.DB)
	eventsRepo := eventsRepository.NewRepository(postgres.DB)
	payrollRepo := payrollRepository.NewRepository(postgres.DB)
	payslipRepo := payslipRepository.NewRepository(postgres.DB)
	reportsRepo := reportsRepository.NewRepository(postgres.DB)
	taskRepo := taskRepository.NewRepository(postgres.DB)

	authSvc := authService.NewAuthService(cognitoSvc, authRepo)
	consentSvc := consentService.NewConsentService(consentRepo)
	orgSvc := organizationService.NewSignUpService(orgRepo, consentRepo, notificationSvc, cognitoSvc, emailSvc)
	employeeSvc := employeeService.NewEmployeeService(employeeRepo)
	userSvc := userService.NewService(cognitoSvc, userRepo)
	eventsSvc := eventsService.NewService(eventsRepo)
	payrollSvc := payrollService.NewPayrollService(payrollRepo, notificationSvc)
	payslipSvc := payslipService.NewPayslipService(payslipRepo, emailSvc, notificationSvc)
	reportsSvc := reportsService.NewReportsService(reportsRepo)
	taskSvc := taskService.NewTaskService(taskRepo)

	if cfg.Anthropic.APIKey == "" {
		logger.Fatal("Anthropic API key is not configured — add [anthropic] api_key to config.toml")
	}
	claudeClient := anthropic.NewClient(cfg.Anthropic.APIKey)
	aiSvc := aiService.NewAIService(claudeClient, taskRepo, reportsSvc)
	aiHTTPHandler := aiHandler.NewAIHandler(aiSvc)
	inviteSvc, err := inviteService.NewService(inviteRepo, notificationSvc, cfg, cognitoClient)
	if err != nil {
		logger.Fatal("Failed to initialize Invite service")
	}

	newAuthHTTPHandler := authHandler.NewAuthHandler(authSvc)
	employeeHTTPHandler := employeeHandler.NewEmployeeHandler(employeeSvc, cognitoSvc)
	handler := organizationHandler.NewOrganizationHandler(orgSvc, consentSvc)
	inviteHTTPHandler := inviteHandler.NewHandler(inviteSvc)
	notificationHTTPHandler := notificationHandler.NewHandler(notificationSvc)
	userHTTPHandler := userHandler.NewHandler(userSvc)
	eventsHTTPHandler := eventsHandler.NewHandler(eventsSvc)
	payrollHTTPHandler := payrollHandler.NewPayrollHandler(payrollSvc)
	payslipHTTPHandler := payslipHandler.NewPayslipHandler(payslipSvc)
	reportsHTTPHandler := reportsHandler.NewReportsHandler(reportsSvc)
	taskHTTPHandler := taskHandler.NewTaskHandler(taskSvc)

	authMw, err := middleware.AuthMiddleware(cfg.AWS.Region, cfg.Cognito.UserPoolID, employeeRepo)
	if err != nil {
		logger.Fatal("Failed to initialize auth middleware")
	}

	router := gin.Default()
	allowedOrigins := []string{
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:5174",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:5174",
	}
	if corsOrigin := os.Getenv("CORS_ORIGIN"); corsOrigin != "" {
		allowedOrigins = append(allowedOrigins, corsOrigin)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders: []string{
			"Content-Disposition",
			"Content-Length",
			"Content-Type",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
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
	v1.POST("/organizations", handler.CreateOrganization)
	v1.POST("/organizations/verify-otp", handler.VerifyOTP)
	v1.POST("/invites/generate", inviteHTTPHandler.GenerateInvite)
	v1.POST("/invites/verify", inviteHTTPHandler.VerifyInvite)
	v1.POST("/invites/complete-registration", inviteHTTPHandler.CompleteRegistration)
	v1.GET("/legal/documents", handler.GetDocuments)
	v1.GET("/cities", cityHTTPHandler.ListCities)

	// Protected routes
	protected := v1.Group("/", authMw)
	adminMw := middleware.RequireAnyRole(middleware.RoleSysAdmin, middleware.RoleAdmin, middleware.RoleHR)
	admin := v1.Group("/", authMw, adminMw)

	// Auth (protected)
	protected.POST("/auth/logout", newAuthHTTPHandler.Logout)

	events := v1.Group("/events")
	events.Use(cognito.AuthMiddleware(cognitoSvc, userRepo))
	events.GET("/upcoming", eventsHTTPHandler.Upcoming)
	events.GET("/my", eventsHTTPHandler.My)
	events.POST("", eventsHTTPHandler.Create)
	events.PATCH("/:id", eventsHTTPHandler.Update)
	events.DELETE("/:id", eventsHTTPHandler.Delete)

	// Consents
	protected.POST("/organizations/consents", handler.SubmitConsents)
	protected.GET("/organizations/consents/validate", handler.ValidateConsents)

	// Notifications
	protected.GET("/notifications", notificationHTTPHandler.ListNotifications)
	protected.PATCH("/notifications/read-all", notificationHTTPHandler.MarkAllAsRead)
	protected.PATCH("/notifications/:id/read", notificationHTTPHandler.MarkAsRead)

	// Departments
	admin.POST("/organizations/departments", handler.CreateDepartment)
	protected.GET("/organizations/departments", handler.ListDepartments)
	admin.DELETE("/organizations/departments/:id", handler.DeleteDepartment)

	// Positions
	admin.POST("/organizations/positions", handler.CreatePosition)
	protected.GET("/organizations/positions", handler.ListPositions)
	admin.DELETE("/organizations/positions/:id", handler.DeletePosition)

	// Employees
	admin.POST("/employees", employeeHTTPHandler.CreateEmployee)
	protected.GET("/employees", employeeHTTPHandler.ListEmployees)
	protected.GET("/employees/:id", employeeHTTPHandler.GetEmployee)
	admin.PATCH("/employees/:id/role", employeeHTTPHandler.UpdateRole)
	admin.PATCH("/employees/:id/salary", employeeHTTPHandler.UpdateSalary)
	admin.PATCH("/employees/:id/status", employeeHTTPHandler.UpdateStatus)
	admin.PATCH("/employees/:id/department", employeeHTTPHandler.UpdateDepartment)
	admin.PATCH("/employees/:id/position", employeeHTTPHandler.UpdatePosition)
	admin.DELETE("/employees/:id", employeeHTTPHandler.DeleteEmployee)

	// Attendance — work schedules
	admin.POST("/attendance/work-schedules", attendanceHTTPHandler.SetWorkSchedule)
	protected.GET("/attendance/work-schedules/:employee_id", attendanceHTTPHandler.GetWorkSchedule)

	// Attendance — СКУД events
	protected.POST("/attendance/skud-events", attendanceHTTPHandler.CreateSkudEvent)
	protected.POST("/attendance/check-in", attendanceHTTPHandler.CheckIn)
	protected.POST("/attendance/check-out", attendanceHTTPHandler.CheckOut)

	// Attendance — leave requests
	protected.POST("/attendance/leave-requests", attendanceHTTPHandler.CreateLeaveRequest)
	protected.GET("/attendance/leave-requests", attendanceHTTPHandler.ListLeaveRequests)
	protected.GET("/attendance/leave-requests/:id", attendanceHTTPHandler.GetLeaveRequest)
	admin.PATCH("/attendance/leave-requests/:id/review", attendanceHTTPHandler.ReviewLeaveRequest)

	// Attendance — records
	protected.GET("/attendance", attendanceHTTPHandler.ListAttendance)
	protected.GET("/attendance/employees/:employee_id", attendanceHTTPHandler.ListEmployeeAttendance)

	// Calendar
	admin.POST("/calendar", calendarHTTPHandler.AddDay)
	admin.DELETE("/calendar/:id", calendarHTTPHandler.DeleteDay)
	protected.GET("/calendar/summary", calendarHTTPHandler.GetMonthSummary)

	// Payroll
	protected.POST("/payroll/preview", payrollHTTPHandler.PreviewPayroll)
	protected.POST("/payroll/cycles", payrollHTTPHandler.CreateCycle)
	protected.GET("/payroll/cycles", payrollHTTPHandler.ListCycles)
	protected.GET("/payroll/cycles/:id", payrollHTTPHandler.GetCycle)
	protected.PATCH("/payroll/cycles/:id/calculate", payrollHTTPHandler.CalculateCycle)
	protected.PATCH("/payroll/cycles/:id/approve", payrollHTTPHandler.ApproveCycle)
	protected.PATCH("/payroll/cycles/:id/prepare-payment", payrollHTTPHandler.PreparePayment)
	protected.PATCH("/payroll/cycles/:id/mark-paid", payrollHTTPHandler.MarkCyclePaid)
	protected.PATCH("/payroll/cycles/:id/reopen", payrollHTTPHandler.ReopenCycle)
	protected.POST("/payroll/cycles/:id/preview", payrollHTTPHandler.PreviewCycle)
	protected.GET("/payroll/cycles/:id/export", payrollHTTPHandler.ExportCycleCSV)
	protected.GET("/payroll/cycles/:id/items", payrollHTTPHandler.ListPayrollItems)
	protected.GET("/payroll/items/:id", payrollHTTPHandler.GetPayrollItem)
	protected.PATCH("/payroll/items/:id/review", payrollHTTPHandler.ReviewPayrollItem)
	protected.POST("/payroll/adjustments", payrollHTTPHandler.CreateAdjustment)
	protected.GET("/payroll/adjustments", payrollHTTPHandler.ListAdjustments)
	protected.DELETE("/payroll/adjustments/:id", payrollHTTPHandler.DeleteAdjustment)
	protected.POST("/payroll/tax-rules", payrollHTTPHandler.CreateTaxRule)
	protected.POST("/payroll/tax-rules/kz-preset", payrollHTTPHandler.CreateKZTaxPreset)
	protected.GET("/payroll/tax-rules", payrollHTTPHandler.ListTaxRules)
	protected.GET("/payroll/policy", payrollHTTPHandler.GetPolicy)
	protected.PUT("/payroll/policy", payrollHTTPHandler.UpsertPolicy)
	protected.POST("/payroll/corrections", payrollHTTPHandler.CreateCorrection)
	protected.POST("/payroll/salary-history", payrollHTTPHandler.CreateSalaryHistory)
	protected.PUT("/payroll/attendance-overrides", payrollHTTPHandler.UpsertAttendanceOverride)
	protected.PATCH("/payroll/employees/:id/employment-period", payrollHTTPHandler.UpdateEmploymentPeriod)

	// Payslips
	protected.POST("/payslips", payslipHTTPHandler.GeneratePayslip)
	protected.GET("/payslips", payslipHTTPHandler.ListPayslips)
	protected.GET("/payslips/:id", payslipHTTPHandler.GetPayslip)
	protected.GET("/payslips/:id/pdf", payslipHTTPHandler.ExportPayslipPDF)
	protected.POST("/payslips/:id/send", payslipHTTPHandler.SendPayslip)
	protected.PATCH("/payslips/:id/void", payslipHTTPHandler.VoidPayslip)
	protected.POST("/payslips/:id/regenerate", payslipHTTPHandler.RegeneratePayslip)
	protected.POST("/payroll/cycles/:id/payslips/generate", payslipHTTPHandler.GenerateCyclePayslips)
	protected.POST("/payroll/cycles/:id/payslips/send", payslipHTTPHandler.SendCyclePayslips)
	protected.GET("/me/payslips", payslipHTTPHandler.ListMyPayslips)
	protected.GET("/me/payslips/:id", payslipHTTPHandler.GetMyPayslip)
	protected.GET("/me/payslips/:id/pdf", payslipHTTPHandler.ExportMyPayslipPDF)

	// Tasks
	adminOnly := v1.Group("/", authMw, middleware.RequireAnyRole(middleware.RoleSysAdmin, middleware.RoleAdmin))
	adminOnly.POST("/tasks", taskHTTPHandler.AssignTask)
	adminOnly.PATCH("/tasks/:id/review", taskHTTPHandler.ReviewTask)
	protected.GET("/tasks", taskHTTPHandler.ListTasks)
	protected.GET("/tasks/:id", taskHTTPHandler.GetTask)
	protected.POST("/tasks/:id/submit", taskHTTPHandler.SubmitReport)

	// AI
	adminOnly.POST("/tasks/:id/ai-review", aiHTTPHandler.ReviewTaskReport)
	adminOnly.GET("/reports/ai-summary", aiHTTPHandler.GetAnalyticsSummary)

	// Reports & Analytics
	protected.GET("/reports/dashboard", reportsHTTPHandler.GetDashboard)
	protected.GET("/reports/payroll", reportsHTTPHandler.GetPayrollSummary)
	protected.GET("/reports/payroll/trends", reportsHTTPHandler.GetPayrollTrends)
	protected.GET("/reports/payroll/departments", reportsHTTPHandler.GetDepartmentPayroll)
	protected.GET("/reports/attendance/today", reportsHTTPHandler.GetAttendanceToday)
	protected.GET("/reports/attendance/weekly", reportsHTTPHandler.GetAttendanceWeekly)
	protected.GET("/reports/employees/statistics", reportsHTTPHandler.GetEmployeeStatistics)
	protected.GET("/reports/departments/statistics", reportsHTTPHandler.GetDepartmentStatistics)
	protected.GET("/reports/export/pdf", reportsHTTPHandler.ExportReportPDF)
	protected.GET("/reports/export/csv", reportsHTTPHandler.ExportReportCSV)

	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":8080"
	}
	logger.Info("Starting HTTP server on port " + port)
	if err := router.Run(port); err != nil {
		logger.Fatal("Server failed", log.Error(err))
	}
}
