package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExecEnv_PinsPath(t *testing.T) {
	t.Setenv("PATH", "/tmp/evil-bin:/usr/bin")
	cmd := exec.Command("true")
	ApplyPinnedExecEnv(cmd, "/proj/bin")

	path := envValue(cmd.Env, "PATH")
	if path == "" {
		t.Fatal("PATH missing from cmd.Env")
	}
	if strings.Contains(path, "/tmp/evil-bin") {
		t.Fatalf("attacker PATH fragment retained: %q", path)
	}
	for _, need := range []string{"/usr/bin", "/bin", "/usr/local/bin", "/proj/bin"} {
		if !pathHas(path, need) {
			t.Fatalf("PATH %q missing allowlisted %q", path, need)
		}
	}
	if runtime.GOOS == "darwin" && !pathHas(path, "/opt/homebrew/bin") {
		t.Fatalf("darwin PATH %q missing /opt/homebrew/bin", path)
	}
}

func TestExecEnv_PreservesCallerExtras(t *testing.T) {
	cmd := exec.Command("true")
	cmd.Env = append(os.Environ(), "DOCKER_IMAGE=example:tag", "PATH=/tmp/evil")
	ApplyPinnedExecEnv(cmd, "")
	if envValue(cmd.Env, "DOCKER_IMAGE") != "example:tag" {
		t.Fatalf("caller DOCKER_IMAGE lost: %v", cmd.Env)
	}
	if strings.Contains(envValue(cmd.Env, "PATH"), "/tmp/evil") {
		t.Fatalf("caller PATH not replaced: %q", envValue(cmd.Env, "PATH"))
	}
}

func TestExecEnv_KeepsHostedToolcacheBeforeUsrBin(t *testing.T) {
	tool := "/opt/hostedtoolcache/go/1.26.5/x64/bin"
	t.Setenv("PATH", "/tmp/evil:"+tool+":/usr/bin")
	cmd := exec.Command("true")
	ApplyPinnedExecEnv(cmd, "/proj/bin")
	path := envValue(cmd.Env, "PATH")
	if !pathHas(path, tool) {
		t.Fatalf("toolcache dir dropped: %q", path)
	}
	if strings.Contains(path, "/tmp/evil") {
		t.Fatalf("evil ambient retained: %q", path)
	}
	toolIdx := pathIndex(path, tool)
	usrIdx := pathIndex(path, "/usr/bin")
	if toolIdx < 0 || usrIdx < 0 || toolIdx > usrIdx {
		t.Fatalf("toolcache must precede /usr/bin: %q", path)
	}
}

func TestPinnedExecPath_ProjectBinFirst(t *testing.T) {
	t.Setenv("PATH", "/usr/bin")
	p := PinnedExecPath(filepath.Join("repo", "bin"))
	parts := strings.Split(p, string(os.PathListSeparator))
	if len(parts) < 1 || parts[0] != filepath.Join("repo", "bin") {
		t.Fatalf("project bin should lead PATH: %q", p)
	}
}

func envValue(env []string, key string) string {
	prefix := key + "="
	for i := len(env) - 1; i >= 0; i-- {
		if strings.HasPrefix(env[i], prefix) {
			return strings.TrimPrefix(env[i], prefix)
		}
	}
	return ""
}

func pathHas(path, dir string) bool {
	return pathIndex(path, dir) >= 0
}

func pathIndex(path, dir string) int {
	for i, p := range strings.Split(path, string(os.PathListSeparator)) {
		if p == dir {
			return i
		}
	}
	return -1
}
