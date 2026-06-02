package validation

import (
	"strconv"
	"strings"

	apivalidation "k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

const maxMQObjectNameLen = 48

// mqReservedPrefixes lists IBM MQ reserved name prefixes (case-insensitive).
// See docs/IBM_MQ_OBJECTS.md and IBM MQ naming reference (SYSTEM.* objects).
var mqReservedPrefixes = []struct {
	prefix  string
	message string
}{
	{prefix: "SYSTEM.", message: "names with prefix SYSTEM. are reserved for queue manager objects"},
	{prefix: "AMQ", message: "names with prefix AMQ are reserved for IBM MQ internal use"},
}

// ValidateKubernetesResourceName checks metadata.name against DNS-1123 subdomain rules.
// An empty name is allowed when the API server assigns one from generateName.
func ValidateKubernetesResourceName(path *field.Path, name string) field.ErrorList {
	if name == "" {
		return nil
	}
	var errs field.ErrorList
	for _, msg := range apivalidation.IsDNS1123Subdomain(name) {
		errs = append(errs, field.Invalid(path, name, msg))
	}
	return errs
}

// ValidateMQObjectName checks IBM MQ object name constraints for queues, topics, channels, and profiles.
func ValidateMQObjectName(path *field.Path, name string) field.ErrorList {
	var errs field.ErrorList

	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return append(errs, field.Required(path, "name must not be empty"))
	}
	if trimmed != name {
		errs = append(errs, field.Invalid(path, name, "name must not have leading or trailing whitespace"))
	}
	upper := strings.ToUpper(trimmed)
	for _, reserved := range mqReservedPrefixes {
		if strings.HasPrefix(upper, reserved.prefix) {
			errs = append(errs, field.Invalid(path, name, reserved.message))
			break
		}
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
