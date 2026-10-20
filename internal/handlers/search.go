package handlers

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/internal/utils"
)

type SearchHandler struct{}

func NewSearchHandler() *SearchHandler {
	return &SearchHandler{}
}

func (h *SearchHandler) SearchUsers(c *gin.Context) {
	name := c.Query("name")

	users, err := searchByName(name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func searchByName(name string) ([]models.User, error) {
	return services.SearchUsersByName(name)
}

func (h *SearchHandler) AdvancedSearch(c *gin.Context) {
	var params models.UserSearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := services.ExecuteSearchQuery(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *SearchHandler) StreamQuery(c *gin.Context) {
	searchTerm := c.Query("q")

	queryBuilder := func() string {
		return "SELECT * FROM items WHERE name LIKE '%" + searchTerm + "%'"
	}

	query := queryBuilder()
	rows, err := services.QueryWithString(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "search completed"})
}

func (h *SearchHandler) DynamicQuery(c *gin.Context) {
	var req struct {
		Table      string                 `json:"table"`
		Conditions map[string]interface{} `json:"conditions"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rows, err := services.DynamicQuery(req.Table, req.Conditions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "query executed"})
}

func (h *SearchHandler) HigherOrderQuery(c *gin.Context) {
	var req struct {
		Field  string `json:"field"`
		Value  string `json:"value"`
		Entity string `json:"entity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	buildFilter := func(field, value string) string {
		return field + " = '" + value + "'"
	}

	filter := buildFilter(req.Field, req.Value)
	rows, err := services.ReflectionBasedQuery(models.User{}, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "query executed"})
}

func (h *SearchHandler) ValidatePattern(c *gin.Context) {
	pattern := c.Query("pattern")
	input := c.Query("input")

	matched, err := regexp.MatchString(pattern, input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pattern"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"matched": matched})
}

func (h *SearchHandler) SearchUsersV2(c *gin.Context) {
	name := c.Query("name")

	if utils.ContainsSQLKeywords(name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid search term"})
		return
	}

	users, err := services.SearchUsersBuilder("username", name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *SearchHandler) AdvancedSearchV2(c *gin.Context) {
	var params models.UserSearchParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 20
	}

	users, err := services.SearchUsersBuilder("department", params.Department)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *SearchHandler) ValidatePatternV2(c *gin.Context) {
	pattern := c.Query("pattern")
	input := c.Query("input")

	if !services.ValidateRegexBounded(pattern) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or dangerous pattern"})
		return
	}

	re, _ := regexp.Compile(pattern)
	matched := re.MatchString(input)

	c.JSON(http.StatusOK, gin.H{"matched": matched})
}

func (h *SearchHandler) FilteredSearch(c *gin.Context) {
	var filters []models.SearchFilter
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	executor := services.NewQueryExecutor("users")
	for _, filter := range filters {
		condition := filter.Field + " " + filter.Operator + " '" + filter.Value.(string) + "'"
		executor.AddFilter(condition)
	}

	rows, err := executor.Execute()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "search completed"})
}

func (h *SearchHandler) FilteredSearchV2(c *gin.Context) {
	var filters []models.SearchFilter
	if err := c.ShouldBindJSON(&filters); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedFields := map[string]bool{"username": true, "email": true, "department": true}
	allowedOperators := map[string]bool{"=": true, "LIKE": true}

	for _, filter := range filters {
		if !allowedFields[filter.Field] || !allowedOperators[filter.Operator] {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filter"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "search completed"})
}

func (h *SearchHandler) TextSearch(c *gin.Context) {
	query := c.Query("q")

	normalizedQuery := utils.NormalizeInput(query)

	sqlQuery := "SELECT * FROM documents WHERE content LIKE '%" + normalizedQuery + "%'"
	rows, err := services.QueryWithString(sqlQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "search completed"})
}

func (h *SearchHandler) TextSearchV2(c *gin.Context) {
	query := c.Query("q")

	rows, err := services.ExecuteNamedQuery(
		"SELECT * FROM documents WHERE content LIKE $1",
		"%"+query+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"message": "search completed"})
}
