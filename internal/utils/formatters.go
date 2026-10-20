package utils

import (
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"
)

func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func FormatCurrency(amount float64, currency string) string {
	return fmt.Sprintf("%s %.2f", currency, amount)
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func FormatPercentage(value, total float64) string {
	if total == 0 {
		return "0%"
	}
	return fmt.Sprintf("%.1f%%", (value/total)*100)
}

func FormatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

type TemplateFormatter struct {
	templates map[string]*template.Template
}

func NewTemplateFormatter() *TemplateFormatter {
	return &TemplateFormatter{
		templates: make(map[string]*template.Template),
	}
}

func (tf *TemplateFormatter) RegisterTemplate(name, tmpl string) error {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return err
	}
	tf.templates[name] = t
	return nil
}

func (tf *TemplateFormatter) Format(name string, data interface{}) (string, error) {
	t, ok := tf.templates[name]
	if !ok {
		return "", fmt.Errorf("template not found: %s", name)
	}

	var buf strings.Builder
	err := t.Execute(&buf, data)
	return buf.String(), err
}

func FormatLogEntry(level, message string, fields map[string]interface{}) string {
	timestamp := time.Now().Format(time.RFC3339)
	var parts []string
	parts = append(parts, fmt.Sprintf("time=%s", timestamp))
	parts = append(parts, fmt.Sprintf("level=%s", level))
	parts = append(parts, fmt.Sprintf("msg=%s", message))

	for key, value := range fields {
		parts = append(parts, fmt.Sprintf("%s=%v", key, value))
	}

	return strings.Join(parts, " ")
}

func FormatError(err error, context string) string {
	return fmt.Sprintf("[%s] Error: %v", context, err)
}

func FormatUserDisplay(username, role string) string {
	return fmt.Sprintf("%s (%s)", username, role)
}

func FormatTableRow(values []string, widths []int) string {
	var parts []string
	for i, v := range values {
		width := 10
		if i < len(widths) {
			width = widths[i]
		}
		parts = append(parts, fmt.Sprintf("%-*s", width, v))
	}
	return strings.Join(parts, " | ")
}

func PadLeft(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return strings.Repeat(pad, length-len(s)) + s
}

func PadRight(s string, length int, pad string) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(pad, length-len(s))
}

func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func FormatList(items []string, separator string) string {
	return strings.Join(items, separator)
}

func FormatKeyValue(key string, value interface{}) string {
	return fmt.Sprintf("%s: %v", key, value)
}

func IntToString(n int) string {
	return strconv.Itoa(n)
}

func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func FormatJSON(data interface{}) (string, error) {
	result, err := MapToJSON(data.(map[string]interface{}))
	return result, err
}

func FormatHeaders(headers map[string]string) string {
	var parts []string
	for key, value := range headers {
		parts = append(parts, fmt.Sprintf("%s: %s", key, value))
	}
	return strings.Join(parts, "\r\n")
}
