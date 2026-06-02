package mqadmin

import "testing"

func TestNormalizeQueueType(t *testing.T) {
	t.Parallel()
	if got := NormalizeQueueType(""); got != QueueTypeLocal {
		t.Fatalf("empty: got %q", got)
	}
	if got := NormalizeQueueType(QueueTypeRemote); got != QueueTypeRemote {
		t.Fatalf("remote: got %q", got)
	}
}
