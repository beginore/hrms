package http

import (
	"errors"
	"hrms/internal/feature/auth/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login godoc
// @Summary      Login
// @Description  Authenticate with email and password, returns JWT tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      service.LoginRequest   true  "Credentials"
// @Success      200   {object}  service.TokenResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailOrPasswordEmpty):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "email and password are required"})
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
		case errors.Is(err, service.ErrUserNotConfirmed):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "account is not confirmed, please verify your email"})
		default:
			log.Printf("[Handler Login] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Logout godoc
// @Summary      Logout
// @Description  Revokes the refresh token and deletes the session
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  service.LogoutRequest  true  "Refresh token"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req service.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.authService.Logout(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, service.ErrRefreshTokenRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			log.Printf("[Handler Logout] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Sends a password reset code to the user's email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  service.ForgotPasswordRequest  true  "Email address"
// @Success      200
// @Failure      400   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req service.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.authService.ForgotPassword(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			log.Printf("[Handler ForgotPassword] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusOK)
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Confirms the password reset using the code sent to the user's email
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body  service.ResetPasswordRequest  true  "Reset details"
// @Success      200
// @Failure      400   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req service.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := h.authService.ResetPassword(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, service.ErrEmailRequired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInvalidResetCode):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrResetCodeExpired):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrPasswordTooWeak):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			log.Printf("[Handler ResetPassword] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusOK)
}

// RefreshTokens godoc
// @Summary      Refresh tokens
// @Description  Exchange a valid refresh token for a new token pair
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      service.RefreshTokenRequest  true  "Refresh token"
// @Success      200   {object}  service.TokenResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var req service.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidRefreshToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
