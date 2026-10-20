package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		statusCode := c.Writer.Status()
		userAgent := c.Request.UserAgent()

		log.Printf("method=%s path=%s status=%d duration=%v ip=%s user_agent=%s",
			method, path, statusCode, duration, clientIP, userAgent)
	}
}

func DetailedLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		username := c.Query("username")
		action := c.Query("action")

		c.Next()

		duration := time.Since(start)
		log.Printf("User %s performed action: %s (took %v)", username, action, duration)
	}
}

func AuditLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		username, _ := c.Get("username")

		resource := c.Request.URL.Path
		method := c.Request.Method
		queryParams := c.Request.URL.RawQuery

		c.Next()

		statusCode := c.Writer.Status()
		log.Printf("AUDIT: user_id=%v username=%v method=%s resource=%s query=%s status=%d",
			userID, username, method, resource, queryParams, statusCode)
	}
}

func ErrorLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Printf("ERROR: %v - Request: %s %s - User-Agent: %s",
					err.Error(), c.Request.Method, c.Request.URL.Path, c.Request.UserAgent())
			}
		}
	}
}

func DebugLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			headers[key] = values[0]
		}

		log.Printf("DEBUG Request: method=%s path=%s headers=%v",
			c.Request.Method, c.Request.URL.Path, headers)

		c.Next()

		log.Printf("DEBUG Response: status=%d", c.Writer.Status())
	}
}

type StructuredLogger struct {
	serviceName string
}

func NewStructuredLogger(serviceName string) *StructuredLogger {
	return &StructuredLogger{serviceName: serviceName}
}

func (l *StructuredLogger) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		entry := fmt.Sprintf(`{"service":"%s","method":"%s","path":"%s","status":%d,"duration_ms":%d}`,
			l.serviceName,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration.Milliseconds())

		log.Println(entry)
	}
}

func MinimalRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		log.Printf("method=%s path=%s status=%d duration=%v",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration)
	}
}
