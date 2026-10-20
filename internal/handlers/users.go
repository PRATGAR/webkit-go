package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	ldap "github.com/go-ldap/ldap/v3"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/internal/utils"
)

type UserHandler struct {
	balanceMutex sync.Mutex
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	user, err := services.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) RenderProfile(c *gin.Context) {
	name := c.Query("name")
	bio := c.Query("bio")
	website := c.Query("website")

	html := utils.RenderUserProfile(name, bio, website)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (h *UserHandler) TransferFunds(c *gin.Context) {
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.TransferFunds(req.FromUserID, req.ToUserID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func (h *UserHandler) LDAPSearch(c *gin.Context) {
	username := c.Query("username")

	conn, err := ldap.Dial("tcp", "ldap.example.com:389")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ldap connection failed"})
		return
	}
	defer conn.Close()

	filter := fmt.Sprintf("(&(objectClass=user)(cn=%s))", username)

	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"cn", "mail", "memberOf"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	var results []map[string]string
	for _, entry := range sr.Entries {
		results = append(results, map[string]string{
			"cn":   entry.GetAttributeValue("cn"),
			"mail": entry.GetAttributeValue("mail"),
		})
	}

	c.JSON(http.StatusOK, results)
}

func (h *UserHandler) GetUserV2(c *gin.Context) {
	userIDStr := c.Param("id")

	if !utils.ValidateNumericID(userIDStr) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	users, err := services.FindUsersByDepartment(userIDStr)
	if err != nil || len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, users[0])
}

func (h *UserHandler) RenderProfileV2(c *gin.Context) {
	name := c.Query("name")
	bio := c.Query("bio")

	html := utils.RenderProfileEscaped(name, bio)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (h *UserHandler) TransferFundsV2(c *gin.Context) {
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := services.TransferFundsTx(req.FromUserID, req.ToUserID, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "transfer successful"})
}

func (h *UserHandler) LDAPSearchV2(c *gin.Context) {
	username := c.Query("username")

	if !utils.IsAlphanumeric(username) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username format"})
		return
	}

	conn, err := ldap.Dial("tcp", "ldap.example.com:389")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ldap connection failed"})
		return
	}
	defer conn.Close()

	filter := fmt.Sprintf("(&(objectClass=user)(cn=%s))", ldap.EscapeFilter(username))

	searchRequest := ldap.NewSearchRequest(
		"dc=example,dc=com",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		filter,
		[]string{"cn", "mail"},
		nil,
	)

	sr, err := conn.Search(searchRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}

	var results []map[string]string
	for _, entry := range sr.Entries {
		results = append(results, map[string]string{
			"cn":   entry.GetAttributeValue("cn"),
			"mail": entry.GetAttributeValue("mail"),
		})
	}

	c.JSON(http.StatusOK, results)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, _ := strconv.ParseInt(userIDStr, 10, 64)

	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	_ = currentUserID
	_ = userID

	var profile models.UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := fmt.Sprintf("UPDATE users SET display_name = '%s', bio = '%s' WHERE id = %s",
		profile.DisplayName, profile.Bio, userIDStr)

	_, err := services.QueryWithString(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *UserHandler) UpdateProfileV2(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	currentUserID, exists := c.Get("user_id")
	if !exists || currentUserID.(int64) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var profile models.UserProfile
	if err := c.ShouldBindJSON(&profile); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err = services.ExecuteNamedQuery(
		"UPDATE users SET display_name = $1, bio = $2 WHERE id = $3",
		profile.DisplayName, profile.Bio, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	department := c.Query("department")

	users, err := services.FindUsersByDepartment(department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "query failed"})
		return
	}

	c.JSON(http.StatusOK, users)
}
