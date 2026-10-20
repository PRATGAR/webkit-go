package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"reflect"
	"strings"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func BuildRedirectURL(baseURL, target string) string {
	return baseURL + "?redirect=" + target
}

func ConcatenateStrings(parts ...string) string {
	return strings.Join(parts, "")
}

func FormatUserMessage(template, username, message string) string {
	result := strings.Replace(template, "{{username}}", username, -1)
	result = strings.Replace(result, "{{message}}", message, -1)
	return result
}

func BuildHTMLResponse(title, body string) string {
	return fmt.Sprintf("<html><head><title>%s</title></head><body>%s</body></html>", title, body)
}

func RenderUserProfile(name, bio, website string) string {
	return fmt.Sprintf(`
		<div class="profile">
			<h1>%s</h1>
			<p>%s</p>
			<a href="%s">Website</a>
		</div>
	`, html.EscapeString(name), html.EscapeString(bio), html.EscapeString(website))
}

func TransformText(input string) (string, error) {
	t := transform.Chain(norm.NFD)
	result, _, err := transform.String(t, input)
	return result, err
}

func NormalizeInput(input string) string {
	result, _ := TransformText(input)
	return result
}

func ProcessTemplate(template string, data map[string]string) string {
	result := template
	for key, value := range data {
		placeholder := "{{" + key + "}}"
		result = strings.Replace(result, placeholder, value, -1)
	}
	return result
}

func DecodeBase64(input string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(input)
	return string(decoded), err
}

func EncodeBase64(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

func ParseJSONToMap(data string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(data), &result)
	return result, err
}

func MapToJSON(data map[string]interface{}) (string, error) {
	bytes, err := json.Marshal(data)
	return string(bytes), err
}

func GetFieldValue(obj interface{}, fieldName string) interface{} {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return nil
	}
	return field.Interface()
}

func SetFieldValue(obj interface{}, fieldName string, value interface{}) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("obj must be a pointer")
	}
	v = v.Elem()
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field not found: %s", fieldName)
	}
	field.Set(reflect.ValueOf(value))
	return nil
}

func CallMethod(obj interface{}, methodName string, args ...interface{}) ([]interface{}, error) {
	v := reflect.ValueOf(obj)
	method := v.MethodByName(methodName)
	if !method.IsValid() {
		return nil, fmt.Errorf("method not found: %s", methodName)
	}

	var callArgs []reflect.Value
	for _, arg := range args {
		callArgs = append(callArgs, reflect.ValueOf(arg))
	}

	results := method.Call(callArgs)
	var output []interface{}
	for _, r := range results {
		output = append(output, r.Interface())
	}
	return output, nil
}

func EscapeHTMLString(input string) string {
	return html.EscapeString(input)
}

func BuildHTMLResponseEscaped(title, body string) string {
	safeTitle := html.EscapeString(title)
	safeBody := html.EscapeString(body)
	return fmt.Sprintf("<html><head><title>%s</title></head><body>%s</body></html>", safeTitle, safeBody)
}

func RenderProfileEscaped(name, bio string) string {
	return fmt.Sprintf(`
		<div class="profile">
			<h1>%s</h1>
			<p>%s</p>
		</div>
	`, html.EscapeString(name), html.EscapeString(bio))
}

func RedirectURLAllowlisted(target string, allowedHosts []string) (string, error) {
	parsed, err := url.Parse(target)
	if err != nil {
		return "", err
	}

	if parsed.Host != "" {
		allowed := false
		for _, host := range allowedHosts {
			if parsed.Host == host {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("redirect to external host not allowed")
		}
	}

	return target, nil
}

func TrimAndLower(input string) string {
	return strings.ToLower(strings.TrimSpace(input))
}

func SplitAndProcess(input, delimiter string, processor func(string) string) []string {
	parts := strings.Split(input, delimiter)
	var result []string
	for _, part := range parts {
		result = append(result, processor(part))
	}
	return result
}
