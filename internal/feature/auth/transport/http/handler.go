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
// POST /auth/login
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

// RefreshTokens godoc
// POST /auth/refresh
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
