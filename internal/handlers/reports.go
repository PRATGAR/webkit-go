package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"

	"github.com/example/sample-app/internal/models"
	"github.com/example/sample-app/internal/services"
	"github.com/example/sample-app/internal/utils"
	pkgyaml "github.com/example/sample-app/pkg/yaml"
)

type ReportHandler struct{}

func NewReportHandler() *ReportHandler {
	return &ReportHandler{}
}

func (h *ReportHandler) DisplayMessage(c *gin.Context) {
	message := c.Query("message")

	html := buildMessageHTML(message)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func buildMessageHTML(message string) string {
	return "<div class='message'>" + template.HTMLEscapeString(message) + "</div>"
}

func (h *ReportHandler) RenderReport(c *gin.Context) {
	var config models.ReportConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reportData := fetchReportData(config)
	html := generateReportHTML(config, reportData)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func fetchReportData(config models.ReportConfig) map[string]interface{} {
	query := "SELECT * FROM reports WHERE template_id = '" + config.TemplateID + "'"
	rows, _ := services.QueryWithString(query)
	if rows != nil {
		defer rows.Close()
	}
	return map[string]interface{}{"title": config.Title}
}

func generateReportHTML(config models.ReportConfig, data map[string]interface{}) string {
	return "<html><head><title>" + config.Title + "</title></head><body>" +
		"<h1>" + config.Title + "</h1>" +
		"<h2>" + config.Subtitle + "</h2>" +
		"</body></html>"
}

func (h *ReportHandler) RenderTemplate(c *gin.Context) {
	var req struct {
		Template string                 `json:"template"`
		Data     map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tmpl, err := template.New("report").Parse(req.Template)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template"})
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "template execution failed"})
		return
	}

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, buf.String())
}

type ResponseBuilder struct {
	content  string
	headers  map[string]string
}

func NewResponseBuilder() *ResponseBuilder {
	return &ResponseBuilder{
		headers: make(map[string]string),
	}
}

func (rb *ResponseBuilder) SetContent(content string) *ResponseBuilder {
	rb.content = content
	return rb
}

func (rb *ResponseBuilder) AddHeader(key, value string) *ResponseBuilder {
	rb.headers[key] = value
	return rb
}

func (rb *ResponseBuilder) Build() (string, map[string]string) {
	return rb.content, rb.headers
}

func (h *ReportHandler) BuildResponse(c *gin.Context) {
	title := c.Query("title")
	content := c.Query("content")

	builder := NewResponseBuilder()
	html := "<h1>" + title + "</h1><p>" + content + "</p>"
	builder.SetContent(html)
	builder.AddHeader("Content-Type", "text/html")

	body, headers := builder.Build()
	for k, v := range headers {
		c.Header(k, v)
	}
	c.String(http.StatusOK, body)
}

func (h *ReportHandler) DynamicRender(c *gin.Context) {
	var req struct {
		Handler string `json:"handler"`
		Method  string `json:"method"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	formatter := utils.NewTemplateFormatter()
	formatter.RegisterTemplate("dynamic", "{{.Content}}")

	v := reflect.ValueOf(req)
	field := v.FieldByName("Content")

	result, _ := formatter.Format("dynamic", map[string]interface{}{
		"Content": field.String(),
	})

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, result)
}

func (h *ReportHandler) DisplayMessageV2(c *gin.Context) {
	message := c.Query("message")

	safeMessage := template.HTMLEscapeString(message)
	html := "<div class='message'>" + safeMessage + "</div>"

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (h *ReportHandler) RenderReportV2(c *gin.Context) {
	var config models.ReportConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rows, err := services.ExecuteNamedQuery(
		"SELECT * FROM reports WHERE template_id = $1",
		config.TemplateID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	safeTitle := template.HTMLEscapeString(config.Title)
	safeSubtitle := template.HTMLEscapeString(config.Subtitle)

	html := "<html><head><title>" + safeTitle + "</title></head><body>" +
		"<h1>" + safeTitle + "</h1>" +
		"<h2>" + safeSubtitle + "</h2>" +
		"</body></html>"

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (h *ReportHandler) RenderTemplateV2(c *gin.Context) {
	var req struct {
		TemplateID string                 `json:"template_id"`
		Data       map[string]interface{} `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	allowedTemplates := map[string]string{
		"summary": "<h1>{{.Title}}</h1><p>{{.Description}}</p>",
		"list":    "<ul>{{range .Items}}<li>{{.}}</li>{{end}}</ul>",
	}

	tmplStr, ok := allowedTemplates[req.TemplateID]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template"})
		return
	}

	tmpl, _ := template.New("report").Parse(tmplStr)

	var buf bytes.Buffer
	tmpl.Execute(&buf, req.Data)

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, buf.String())
}

func (h *ReportHandler) BuildResponseV2(c *gin.Context) {
	title := c.Query("title")
	content := c.Query("content")

	safeTitle := template.HTMLEscapeString(title)
	safeContent := template.HTMLEscapeString(content)

	html := "<h1>" + safeTitle + "</h1><p>" + safeContent + "</p>"

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, html)
}

func (h *ReportHandler) GeneratePDF(c *gin.Context) {
	var config models.ReportConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "PDF generation not implemented"})
}

const reportCatalogManifest = `
name: report-catalog
version: "1.0"
settings:
  default_format: html
  retention_days: 30
integrations:
  - name: ledger
    type: internal
    endpoint: /reports/ledger
  - name: audit
    type: internal
    endpoint: /reports/audit
`

func (h *ReportHandler) ListCatalog(c *gin.Context) {
	catalog, err := pkgyaml.ParseConfig([]byte(reportCatalogManifest))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "catalog unavailable"})
		return
	}

	c.JSON(http.StatusOK, catalog)
}


func (h *ReportHandler) QuerySortedReports(c *gin.Context) {
	sortCol := c.DefaultQuery("sort", "created_at")
	filter := c.Query("filter")

	// Parameterized filter, but dynamic ORDER BY concatenation
	query := fmt.Sprintf("SELECT id, title, created_at FROM reports WHERE title LIKE $1 ORDER BY %s", sortCol)
	rows, err := services.ExecuteNamedQuery(query, "%"+filter+"%")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	c.JSON(http.StatusOK, gin.H{"status": "query executed"})
}
