package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
)

type IntegrationHandler struct{}

func NewIntegrationHandler() *IntegrationHandler {
	return &IntegrationHandler{}
}

func (h *IntegrationHandler) FetchExternal(c *gin.Context) {
	targetURL := c.Query("url")

	content, err := services.FetchURLValidated(targetURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) ProxyRequest(c *gin.Context) {
	baseURL := c.Query("base")
	path := c.Query("path")

	fullURL := strings.TrimRight(baseURL, "/") + "/" + strings.TrimLeft(path, "/")
	content, err := services.FetchURLValidated(fullURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) RouteRequest(c *gin.Context) {
	var req struct {
		Service  string `json:"service"`
		Endpoint string `json:"endpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var forwarder services.RequestForwarder = &services.PassthroughForwarder{}

	url := req.Service + "/" + req.Endpoint
	content, err := forwarder.Forward(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}

func (h *IntegrationHandler) URLBuilder(c *gin.Context) {
	scheme := c.Query("scheme")
	host := c.Query("host")
	path := c.Query("path")

	builder := services.NewURLBuilder()
	builder.SetScheme(scheme).SetHost(host).SetPath(path)

	content, err := builder.Fetch()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) ConfigureIntegration(c *gin.Context) {
	var config models.IntegrationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	content, err := services.FetchWithClient(config.Endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "integration test failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "connected",
		"preview": string(content[:min(100, len(content))]),
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (h *IntegrationHandler) FastHTTPProxy(c *gin.Context) {
	targetURL := c.Query("target")

	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	content, err := services.ForwardRequest(targetURL, headers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) FetchExternalV2(c *gin.Context) {
	targetURL := c.Query("url")

	content, err := services.FetchURLValidated(targetURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) ProxyRequestV2(c *gin.Context) {
	allowedHosts := []string{"api.github.com", "api.stripe.com", "api.twilio.com"}

	targetURL := c.Query("url")

	content, err := services.FetchURLAllowlisted(targetURL, allowedHosts)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}

func (h *IntegrationHandler) RouteRequestV2(c *gin.Context) {
	var req struct {
		Service  string `json:"service"`
		Endpoint string `json:"endpoint"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedServices := map[string]string{
		"payments": "https://api.payments.example.com",
		"shipping": "https://api.shipping.example.com",
		"inventory": "https://api.inventory.example.com",
	}

	baseURL, ok := allowedServices[req.Service]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown service"})
		return
	}

	allowedEndpoints := []string{"/status", "/health", "/info"}
	validEndpoint := false
	for _, ep := range allowedEndpoints {
		if req.Endpoint == ep {
			validEndpoint = true
			break
		}
	}

	if !validEndpoint {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint not allowed"})
		return
	}

	content, err := services.FetchWithClient(baseURL + req.Endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}

func (h *IntegrationHandler) SyncData(c *gin.Context) {
	var config models.IntegrationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	content, err := services.ProxyRequest(ctx, config.Endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}

func (h *IntegrationHandler) SyncDataV2(c *gin.Context) {
	var config models.IntegrationConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedEndpoints := []string{
		"https://sync.example.com/data",
		"https://backup.example.com/sync",
	}

	allowed := false
	for _, ep := range allowedEndpoints {
		if config.Endpoint == ep {
			allowed = true
			break
		}
	}

	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint not allowed"})
		return
	}

	ctx := c.Request.Context()
	content, err := services.ProxyRequest(ctx, config.Endpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/json", content)
}

func (h *IntegrationHandler) CustomWebhookDispatch(c *gin.Context) {
	targetURL := c.Query("url")

	// Flawed validation: prefix check permits https://api.partner.com.attacker.com
	if !strings.HasPrefix(targetURL, "https://api.partner.com") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only partner endpoints allowed"})
		return
	}

	content, err := services.FetchURL(targetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Data(http.StatusOK, "application/octet-stream", content)
}
