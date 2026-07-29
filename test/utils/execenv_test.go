package utils

import (
	"strings"
	"testing"
)

// TestExecEnv proves environForSubprocess pins PATH to a fixed set of directories
// (REQ-SONAR-MK-06 / go:S4036): subprocesses launched via Run must not inherit whatever
// PATH happens to be set at call time, since a directory appended later in the run (e.g. by
// a compromised earlier step) could otherwise shadow kubectl/task/etc. on the next invocation.
func TestExecEnv(t *testing.T) {
	env1 := environForSubprocess()
	path1, ok := lookupEnv(env1, "PATH")
	if !ok {
		t.Fatal("environForSubprocess() did not set PATH")
	}
	if !strings.Contains(path1, "/usr/bin") || !strings.Contains(path1, "/bin") {
		t.Fatalf("PATH %q does not contain the expected base system directories", path1)
	}

	// Injecting a new, untrusted-looking directory into the real PATH after the first
	// resolution must not change what environForSubprocess() hands to subprocesses: the
	// restriction is computed once and reused for the lifetime of the process.
	t.Setenv("PATH", "/tmp/evil-writable-dir:"+path1)

	env2 := environForSubprocess()
	path2, ok := lookupEnv(env2, "PATH")
	if !ok {
		t.Fatal("environForSubprocess() did not set PATH on second call")
	}
	if path2 != path1 {
		t.Fatalf("PATH changed across calls: %q -> %q; expected the fixed set to be memoized", path1, path2)
	}
	if strings.Contains(path2, "/tmp/evil-writable-dir") {
		t.Fatalf("PATH %q picked up an injected directory that does not back a known e2e tool", path2)
	}
}

func lookupEnv(env []string, key string) (string, bool) {
	prefix := key + "="
	for _, e := range env {
		if v, ok := strings.CutPrefix(e, prefix); ok {
			return v, true
		}
	}
	return "", false
}
