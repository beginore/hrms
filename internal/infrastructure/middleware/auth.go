package middleware

import (
	"fmt"
	"net/http"
	"strings"

	mkeyfunc "github.com/MicahParks/keyfunc/v3"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	employeeRepo "hrms/internal/feature/employee/repository"
)

const (
	OrgIDKey     = "orgID"
	UserIDKey    = "userID"
	RoleKey      = "role"
	RoleSysAdmin = "SysAdmin"
)

func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(RoleKey)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		userRoleStr, ok := userRole.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid role type"})
			return
		}
		if userRoleStr != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

func AuthMiddlewareWithKeyfunc(
	kf jwt.Keyfunc,
	empRepo employeeRepo.EmployeeRepository,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenStr, kf)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		userIDStr, _ := claims["sub"].(string)
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid userID in token"})
			return
		}

		userInfo, err := empRepo.GetRoleAndOrgIDByUserID(c.Request.Context(), userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to fetch user info"})
			return
		}

		c.Set(UserIDKey, userID)
		c.Set(OrgIDKey, userInfo.OrgID)
		c.Set(RoleKey, userInfo.Role)

		c.Next()
	}
}

func AuthMiddleware(region, userPoolID string, empRepo employeeRepo.EmployeeRepository) (gin.HandlerFunc, error) {
	jwksURL := fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		region,
		userPoolID,
	)

	k, err := mkeyfunc.NewDefault([]string{jwksURL})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	return AuthMiddlewareWithKeyfunc(k.Keyfunc, empRepo), nil
}
