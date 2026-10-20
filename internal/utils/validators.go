package utils

import (
	"net/mail"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func ValidateUsername(username string) bool {
	if len(username) < 3 || len(username) > 50 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}

func ValidatePassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

func ValidateNumericID(id string) bool {
	_, err := strconv.ParseInt(id, 10, 64)
	return err == nil
}

func ValidateFilename(filename string) bool {
	if strings.Contains(filename, "..") {
		return false
	}
	if strings.ContainsAny(filename, "/\\") {
		return false
	}
	if len(filename) > 255 {
		return false
	}
	return true
}

func ValidatePath(path string) bool {
	if strings.Contains(path, "..") {
		return false
	}
	return true
}

func ValidateURL(urlStr string) bool {
	matched, _ := regexp.MatchString(`^https?://[a-zA-Z0-9.-]+(/.*)?$`, urlStr)
	return matched
}

func ValidateRole(role string) bool {
	validRoles := map[string]bool{
		"admin":    true,
		"user":     true,
		"guest":    true,
		"moderator": true,
	}
	return validRoles[role]
}

func ValidateAction(action string) bool {
	validActions := []string{"view", "edit", "delete", "create", "list"}
	for _, a := range validActions {
		if action == a {
			return true
		}
	}
	return false
}

func SanitizeFilename(filename string) string {
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	return filename
}

func SanitizeInput(input string) string {
	dangerous := []string{"<script>", "</script>", "javascript:", "onerror=", "onclick="}
	result := input
	for _, d := range dangerous {
		result = strings.ReplaceAll(strings.ToLower(result), d, "")
	}
	return result
}

func IsAlphanumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func IsNumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return len(s) > 0
}

func ContainsSQLKeywords(input string) bool {
	keywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "DROP", "UNION", "--", ";"}
	upper := strings.ToUpper(input)
	for _, keyword := range keywords {
		if strings.Contains(upper, keyword) {
			return true
		}
	}
	return false
}

func ValidateJSONStructure(data map[string]interface{}, requiredFields []string) bool {
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			return false
		}
	}
	return true
}

func ValidateContentType(contentType string, allowed []string) bool {
	for _, a := range allowed {
		if contentType == a {
			return true
		}
	}
	return false
}

func ValidateFileExtension(filename string, allowed []string) bool {
	for _, ext := range allowed {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}
	return false
}

func ValidateIPAddress(ip string) bool {
	matched, _ := regexp.MatchString(`^(\d{1,3}\.){3}\d{1,3}$`, ip)
	return matched
}
