package mqrest

import (
	"strings"
	"testing"
)

// FuzzParseMQSCDisplayAttributes fuzzes the parser that turns untrusted mqweb REST
// runCommand DISPLAY text into KEY(value) attribute maps. The text originates from an
// external queue manager, so the parser must never panic and must uphold its output
// contract on arbitrary bytes. (Replaces the conversion round-trip fuzz retired with
// v1alpha1 in Phase 8e; see CI-20 / OpenSSF Scorecard "Fuzzing".)
func FuzzParseMQSCDisplayAttributes(f *testing.F) {
	seeds := []string{
		"AMQ8864I: Display authority record details. PROFILE(APP.Q) OBJTYPE(QUEUE) AUTHLIST(GET)",
		"AMQ8878I: Display channel authentication record details.   CHLAUTH(SYSTEM.*) TYPE(BLOCKUSER)",
		"details QUEUE(DEV.QUEUE.1) MAXDEPTH(5000) CURDEPTH(0)",
		"details. DESCR('a value with spaces') USAGE(NORMAL)",
		"AMQ8409I: Display Queue details.\nQUEUE(APP.Q) TYPE(QLOCAL)\nMAXDEPTH(100)",
		"no marker here KEY(v)",
		"details. K()",
		"details. LOWER(x) key(ignored)",
		"",
		"()",
		"details.",
		"'''",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		// Mirror runCommand text output, which arrives as separate lines.
		lines := strings.Split(raw, "\n")

		attrs := parseMQSCDisplayAttributes(lines) // contract: never panics

		for k, v := range attrs {
			// Keys come from the regex group [A-Z][A-Z0-9]* lowercased, so they are
			// non-empty and drawn only from [a-z0-9].
			if k == "" {
				t.Fatalf("empty attribute key (value %q, raw %q)", v, raw)
			}
			if k != strings.ToLower(k) {
				t.Fatalf("key %q is not lower-cased (raw %q)", k, raw)
			}
			for _, r := range k {
				isLower := r >= 'a' && r <= 'z'
				isDigit := r >= '0' && r <= '9'
				if !isLower && !isDigit {
					t.Fatalf("key %q contains unexpected rune %q (raw %q)", k, r, raw)
				}
			}
			// Values are Trim'd of surrounding single quotes, so no value may start or
			// end with one. (Interior quotes and interior padding are preserved.)
			if strings.HasPrefix(v, "'") || strings.HasSuffix(v, "'") {
				t.Fatalf("value %q retains a surrounding single quote (key %q, raw %q)", v, k, raw)
			}
		}
	})
}
