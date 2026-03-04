package util

import (
	"regexp"
	"strings"
)

// Validator 验证器
type Validator struct {
	emailRegex    *regexp.Regexp
	apiKeyRegex   *regexp.Regexp
	modelNameRegex *regexp.Regexp
}

// NewValidator 创建新的验证器
func NewValidator() *Validator {
	return &Validator{
		emailRegex:     regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`),
		apiKeyRegex:    regexp.MustCompile(`^sk-[a-zA-Z0-9]{20,}$`),
		modelNameRegex: regexp.MustCompile(`^[a-z0-9-]+$`),
	}
}

// ValidateEmail 验证邮箱格式
func (v *Validator) ValidateEmail(email string) bool {
	return v.emailRegex.MatchString(email)
}

// ValidateAPIKey 验证 API Key 格式
func (v *Validator) ValidateAPIKey(key string) bool {
	return v.apiKeyRegex.MatchString(key)
}

// ValidateModelName 验证模型名称格式
func (v *Validator) ValidateModelName(name string) bool {
	return v.modelNameRegex.MatchString(name)
}

// ValidateURL 验证 URL 格式
func (v *Validator) ValidateURL(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

// ValidateRequired 验证必填字段
func (v *Validator) ValidateRequired(value string) bool {
	return strings.TrimSpace(value) != ""
}

// ValidateLength 验证字符串长度
func (v *Validator) ValidateLength(value string, min, max int) bool {
	length := len(value)
	return length >= min && length <= max
}

// ValidatePositive 验证正数
func (v *Validator) ValidatePositive(value int64) bool {
	return value > 0
}

// ValidatePositiveFloat 验证正浮点数
func (v *Validator) ValidatePositiveFloat(value float64) bool {
	return value > 0
}
