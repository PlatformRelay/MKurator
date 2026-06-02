package mqrest

import (
	"regexp"
	"strings"
)

var mqscDisplayPair = regexp.MustCompile(`([A-Z][A-Z0-9]*)\(([^)]*)\)`)

// parseMQSCDisplayAttributes extracts KEY(value) pairs from runCommand DISPLAY text lines.
func parseMQSCDisplayAttributes(lines []string) map[string]string {
	attrs := make(map[string]string)
	combined := strings.Join(lines, " ")
	for _, marker := range []string{"details.", "details "} {
		if idx := strings.Index(combined, marker); idx >= 0 {
			combined = strings.TrimSpace(combined[idx+len(marker):])
			break
		}
	}
	for _, m := range mqscDisplayPair.FindAllStringSubmatch(combined, -1) {
		key := strings.ToLower(m[1])
		val := strings.Trim(strings.TrimSpace(m[2]), "'")
		attrs[key] = val
	}
	return attrs
}

func (r *mqscResponse) displayTextAttributes() map[string]string {
	var lines []string
	for _, cr := range r.CommandResponse {
		lines = append(lines, cr.Text...)
		lines = append(lines, cr.Message...)
	}
	return parseMQSCDisplayAttributes(lines)
}
