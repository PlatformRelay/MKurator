package validation

import (
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

const maxMQObjectNameLen = 48

// ValidateMQObjectName checks IBM MQ object name constraints for queues, topics, and channels.
func ValidateMQObjectName(path *field.Path, name string) field.ErrorList {
	var errs field.ErrorList

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return append(errs, field.Required(path, "name must not be empty"))
	}
	if trimmed != name {
		errs = append(errs, field.Invalid(path, name, "name must not have leading or trailing whitespace"))
	}
	if strings.HasPrefix(strings.ToUpper(trimmed), "SYSTEM.") {
		errs = append(errs, field.Invalid(path, name, "names with prefix SYSTEM. are reserved"))
	}
	if strings.HasPrefix(trimmed, ".") || strings.HasSuffix(trimmed, ".") {
		errs = append(errs, field.Invalid(path, name, "name must not start or end with '.'"))
	}
	if len(trimmed) > maxMQObjectNameLen {
		errs = append(errs, field.Invalid(path, name, "name must be at most 48 characters"))
	}
	for i, r := range trimmed {
		if !isAllowedMQNameRune(r) {
			errs = append(errs, field.Invalid(path, name,
				"name contains invalid character at position "+strconv.Itoa(i)+
					"; allowed: A-Z, 0-9, ., /, %, &, $, #, @"))
			break
		}
	}
	return errs
}

func isAllowedMQNameRune(r rune) bool {
	switch {
	case r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	case r == '.', r == '/', r == '%', r == '&', r == '$', r == '#', r == '@':
		return true
	default:
		return false
	}
}
