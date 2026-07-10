package validation

import "testing"

func TestAttrValue(t *testing.T) {
	t.Parallel()
	attrs := map[string]string{"conname": "host(1414)", "topstr": "orders/#"}
	cases := []struct {
		name  string
		attrs map[string]string
		key   string
		want  string
	}{
		{"nil map", nil, "conname", ""},
		{"direct hit", attrs, "conname", "host(1414)"},
		{"normalized case-fold hit", attrs, "CONNAME", "host(1414)"},
		{"normalized topicstr alias hit", attrs, "topicstr", "orders/#"},
		{"miss", attrs, "xmitq", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := attrValue(tc.attrs, tc.key); got != tc.want {
				t.Fatalf("attrValue(%v, %q) = %q, want %q", tc.attrs, tc.key, got, tc.want)
			}
		})
	}
}
