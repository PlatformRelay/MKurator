package validation

import (
	"strings"
	"testing"
)

func TestDeprecatedQueueAttributeWarnings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		attrs    map[string]string
		contains []string
	}{
		{
			name: "queue promoted keys",
			attrs: map[string]string{
				"maxdepth": "5000",
				"targq":    "APP.Q",
			},
			contains: []string{"maxdepth", "spec.maxDepth", "targq", "spec.targetQueue"},
		},
		{
			name: "queue normalized key",
			attrs: map[string]string{
				"REMOTEmanager": "QM2",
			},
			contains: []string{"REMOTEmanager", "spec.remoteQueueManager"},
		},
		{
			name:     "queue no deprecated keys",
			attrs:    map[string]string{"custom": "value"},
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			warnings := deprecatedQueueAttributeWarnings(tt.attrs)
			assertWarningContains(t, warnings, tt.contains...)
		})
	}
}

func TestDeprecatedTopicAttributeWarnings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		attrs    map[string]string
		contains []string
	}{
		{
			name: "topic promoted keys",
			attrs: map[string]string{
				"topicstr": "retail/orders",
				"pub":      "enabled",
			},
			contains: []string{"topicstr", "spec.topicString", "pub", "spec.publish"},
		},
		{
			name: "topic no deprecated keys",
			attrs: map[string]string{
				"cluster": "MYCLUSTER",
			},
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			warnings := deprecatedTopicAttributeWarnings(tt.attrs)
			assertWarningContains(t, warnings, tt.contains...)
		})
	}
}

func TestDeprecatedChannelAttributeWarnings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		attrs    map[string]string
		contains []string
	}{
		{
			name: "channel promoted keys",
			attrs: map[string]string{
				"maxmsgl": "4194304",
				"conname": "qm2.example.com(1414)",
			},
			contains: []string{"maxmsgl", "spec.maxMsgLength", "conname", "spec.connName"},
		},
		{
			name: "channel normalized key",
			attrs: map[string]string{
				"SSLCAUTH": "required",
			},
			contains: []string{"SSLCAUTH", "spec.sslClientAuth"},
		},
		{
			name: "channel no deprecated keys",
			attrs: map[string]string{
				"custom": "value",
			},
			contains: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			warnings := deprecatedChannelAttributeWarnings(tt.attrs)
			assertWarningContains(t, warnings, tt.contains...)
		})
	}
}

func assertWarningContains(t *testing.T, warnings []string, expected ...string) {
	t.Helper()
	if len(expected) == 0 {
		if len(warnings) != 0 {
			t.Fatalf("expected no warnings, got %v", warnings)
		}
		return
	}
	joined := strings.Join(warnings, " ")
	for _, part := range expected {
		if !strings.Contains(joined, part) {
			t.Fatalf("warnings %q do not contain %q", joined, part)
		}
	}
}
