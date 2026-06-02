package mqrest

import "testing"

func TestTruncate(t *testing.T) {
	t.Parallel()
	if got := truncate("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
	long := "abcdefghijklmnopqrstuvwxyz"
	if got := truncate(long, 10); got != "abcdefghij..." {
		t.Fatalf("got %q", got)
	}
}
