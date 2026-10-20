package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
)

type WebhookHandler struct{}

func NewWebhookHandler() *WebhookHandler {
	return &WebhookHandler{}
}

func (h *WebhookHandler) SetHeader(c *gin.Context) {
	headerName := c.Query("name")
	headerValue := c.Query("value")

	allowedHeaders := map[string]bool{
		"X-Custom-Header": true,
		"X-Trace-ID":      true,
		"X-Webhook-ID":    true,
	}
	if !allowedHeaders[headerName] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "header not allowed"})
		return
	}

	headerValue = strings.ReplaceAll(strings.ReplaceAll(headerValue, "\n", ""), "\r", "")
	c.Header(headerName, headerValue)
	c.JSON(http.StatusOK, gin.H{"message": "header set"})
}

func (h *WebhookHandler) Forward(c *gin.Context) {
	var config models.WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, _ := ioutil.ReadAll(c.Request.Body)

	err := services.MakeWebhookCall(config.URL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "webhook failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook sent"})
}

func (h *WebhookHandler) WebhookForward(c *gin.Context) {
	var req struct {
		TargetURL string            `json:"target_url"`
		Headers   map[string]string `json:"headers"`
		Payload   interface{}       `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	callback := func(url string) ([]byte, error) {
		return services.FetchWithHeaders(url, req.Headers)
	}

	handler := &services.CallbackHandler{}
	handler.SetCallback(callback)

	result, err := handler.Execute(req.TargetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", result)
}

func (h *WebhookHandler) ProcessEvent(c *gin.Context) {
	eventType := c.GetHeader("X-Event-Type")
	signature := c.GetHeader("X-Signature")

	body, _ := ioutil.ReadAll(c.Request.Body)

	_ = signature

	switch eventType {
	case "user.created":
		h.handleUserCreated(c, body)
	case "order.completed":
		h.handleOrderCompleted(c, body)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown event type"})
	}
}

func (h *WebhookHandler) handleUserCreated(c *gin.Context, body []byte) {
	var event struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"processed": true})
}

func (h *WebhookHandler) handleOrderCompleted(c *gin.Context, body []byte) {
	var event struct {
		OrderID    string  `json:"order_id"`
		Amount     float64 `json:"amount"`
		WebhookURL string  `json:"webhook_url"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if event.WebhookURL != "" {
		services.FetchURL(event.WebhookURL)
	}

	c.JSON(http.StatusOK, gin.H{"processed": true})
}

func (h *WebhookHandler) DynamicResponse(c *gin.Context) {
	responseHeader := c.Query("response_header")
	responseValue := c.Query("response_value")

	if responseHeader != "" {
		c.Header(responseHeader, responseValue)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *WebhookHandler) SetHeaderV2(c *gin.Context) {
	headerValue := c.Query("value")

	allowedHeaders := map[string]bool{
		"X-Request-ID":    true,
		"X-Correlation-ID": true,
	}

	headerName := c.Query("name")
	if !allowedHeaders[headerName] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "header not allowed"})
		return
	}

	c.Header(headerName, headerValue)
	c.JSON(http.StatusOK, gin.H{"message": "header set"})
}

func (h *WebhookHandler) ForwardV2(c *gin.Context) {
	var config models.WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedHosts := []string{"api.example.com", "hooks.example.com"}

	body, _ := ioutil.ReadAll(c.Request.Body)

	result, err := services.FetchURLAllowlisted(config.URL, allowedHosts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_ = body
	c.Data(http.StatusOK, "application/json", result)
}

func (h *WebhookHandler) ProcessEventV2(c *gin.Context) {
	eventType := c.GetHeader("X-Event-Type")
	signature := c.GetHeader("X-Signature")

	if signature == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing signature"})
		return
	}

	validEvents := map[string]bool{
		"user.created":    true,
		"order.completed": true,
	}

	if !validEvents[eventType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown event type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"processed": true})
}

func (h *WebhookHandler) RegisterWebhook(c *gin.Context) {
	var config models.WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := services.FetchURL(config.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook URL not reachable"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook registered"})
}

func (h *WebhookHandler) RegisterWebhookV2(c *gin.Context) {
	var config models.WebhookConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	_, err := services.FetchURLValidated(config.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "webhook URL not valid"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "webhook registered"})
}
