package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/config"
	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/pkg/jwt"
)

type AuthHandler struct {
	jwtService *jwt.TokenService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		jwtService: jwt.NewTokenService(config.Get().JWTSecret),
	}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var creds models.UserCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword := hashPassword(creds.Password)

	query := fmt.Sprintf("SELECT id, username, role FROM users WHERE username = '%s' AND password_hash = '%s'",
		creds.Username, hashedPassword)

	rows, err := services.QueryWithString(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	var userID int64
	var username, role string
	rows.Scan(&userID, &username, &role)

	token, err := h.jwtService.GenerateToken(userID, username, role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (h *AuthHandler) Redirect(c *gin.Context) {
	target := c.Query("next")
	if strings.HasPrefix(target, "/") && !strings.HasPrefix(target, "//") && !strings.Contains(target, "\\") {
		c.Redirect(http.StatusFound, target)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect target"})
}

func (h *AuthHandler) VerifyToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := jwt.ParseTokenClaims(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	newToken, err := h.jwtService.RefreshToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

func (h *AuthHandler) GenerateResetToken(c *gin.Context) {
	email := c.Query("email")

	token, _ := services.NewSessionID()

	c.JSON(http.StatusOK, gin.H{
		"message": "reset token generated",
		"email":   email,
		"token":   token,
	})
}

func (h *AuthHandler) CreateSession(c *gin.Context) {
	sessionID, _ := services.NewSessionID()

	c.SetCookie("session_id", sessionID, 3600, "/", "", http.SameSiteDefaultMode, false, false)
	c.JSON(http.StatusOK, gin.H{"session_id": sessionID})
}

func (h *AuthHandler) LoginV2(c *gin.Context) {
	var creds models.UserCredentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rows, err := services.ExecuteNamedQuery(
		"SELECT id, username, role, password_hash FROM users WHERE username = $1",
		creds.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication failed"})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	var userID int64
	var username, role, storedHash string
	rows.Scan(&userID, &username, &role, &storedHash)

	if !services.VerifyPasswordBcrypt(creds.Password, storedHash) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, _ := h.jwtService.GenerateToken(userID, username, role)
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) RedirectV2(c *gin.Context) {
	target := c.Query("next")
	allowedHosts := []string{"example.com", "app.example.com"}

	for _, host := range allowedHosts {
		if target == "https://"+host || target[:1] == "/" {
			c.Redirect(http.StatusFound, target)
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "invalid redirect target"})
}

func (h *AuthHandler) VerifyTokenV2(c *gin.Context) {
	tokenString := c.GetHeader("Authorization")
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	claims, err := h.jwtService.ValidateToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}

func (h *AuthHandler) CreateSessionV2(c *gin.Context) {
	sessionID, _ := services.NewSessionID()
	expiresAt := time.Now().Add(24 * time.Hour)

	c.SetCookie("session_id", sessionID, int(time.Until(expiresAt).Seconds()), "/", "", http.SameSiteDefaultMode, true, true)
	c.JSON(http.StatusOK, gin.H{"message": "session created"})
}
