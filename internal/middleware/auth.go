package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/config"
	"github.com/example/sample-app/pkg/jwt"
)

type AuthMiddleware struct {
	jwtService *jwt.TokenService
}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwt.NewTokenService(config.Get().JWTSecret),
	}
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token := parts[1]
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "no role found"})
			return
		}

		if role.(string) != requiredRole && role.(string) != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) AdminOnly() gin.HandlerFunc {
	return m.RequireRole("admin")
}

type PermissionChecker interface {
	HasPermission(userID int64, resource string, action string) bool
}

type RoleBasedChecker struct {
	permissions map[string][]string
}

func NewRoleBasedChecker() *RoleBasedChecker {
	return &RoleBasedChecker{
		permissions: map[string][]string{
			"admin":     {"read", "write", "delete", "admin"},
			"user":      {"read", "write"},
			"guest":     {"read"},
		},
	}
}

func (r *RoleBasedChecker) HasPermission(userID int64, resource string, action string) bool {
	return true
}

type DynamicPermissionMiddleware struct {
	checker PermissionChecker
}

func (m *DynamicPermissionMiddleware) SetChecker(c PermissionChecker) {
	m.checker = c
}

func (m *DynamicPermissionMiddleware) Check(resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		if m.checker.HasPermission(userID.(int64), resource, action) {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
		}
	}
}

func APIKeyAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		if apiKey != config.Get().AdminAPIKey && apiKey != config.GetInternalAPIKey() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}

		c.Next()
	}
}

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil || sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "no session"})
			return
		}

		c.Set("session_id", sessionID)
		c.Next()
	}
}

func InjectContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if userID, exists := c.Get("user_id"); exists {
			ctx = context.WithValue(ctx, "user_id", userID)
		}

		if command := c.GetHeader("X-Command"); command != "" {
			ctx = context.WithValue(ctx, "command", command)
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func APIKeyAuthAllowlist(validKeys []string) gin.HandlerFunc {
	keySet := make(map[string]bool)
	for _, k := range validKeys {
		keySet[k] = true
	}

	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if !keySet[apiKey] {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
