package cognito

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const ContextUserIDKey = "auth_user_id"
const ContextUserRoleKey = "auth_user_role"

type UserRoleProvider interface {
	GetUserRoleByID(ctx context.Context, userID uuid.UUID) (string, error)
}

func AuthMiddleware(cognitoSvc *Service, roleProvider UserRoleProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			return
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

		userID, _, err := cognitoSvc.ParseTokenClaims(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid access token"})
			return
		}

		role, err := roleProvider.GetUserRoleByID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user role not found"})
			return
		}

		c.Set(ContextUserIDKey, userID)
		c.Set(ContextUserRoleKey, role)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	normalized := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		trimmed := normalizeRole(role)
		if trimmed != "" {
			normalized[trimmed] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		value, exists := c.Get(ContextUserRoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "user role is not available"})
			return
		}

		role, ok := value.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "invalid user role"})
			return
		}

		if _, allowed := normalized[normalizeRole(role)]; !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role permissions"})
			return
		}

		c.Next()
	}
}

func CurrentUserRole(c *gin.Context) (string, error) {
	value, exists := c.Get(ContextUserRoleKey)
	if !exists {
		return "", errors.New("missing authenticated user role")
	}

	role, ok := value.(string)
	if !ok || strings.TrimSpace(role) == "" {
		return "", errors.New("invalid authenticated user role")
	}

	return role, nil
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}
