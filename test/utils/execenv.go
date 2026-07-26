package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PinnedExecPath returns the allowlisted PATH for test subprocesses (Sonar go:S4036).
//
// Order matters: repo bin/ and CI toolchains (hostedtoolcache / setup-go) come before
// /usr/bin so `go`/`task` resolve to the job's toolchain, not a stale system Go.
// Ambient PATH entries outside the allowlist (e.g. /tmp/evil-bin) are dropped.
func PinnedExecPath(projectBin string) string {
	seen := make(map[string]struct{}, 16)
	parts := make([]string, 0, 16)
	add := func(p string) {
		if p == "" {
			return
		}
		if _, ok := seen[p]; ok {
			return
		}
		seen[p] = struct{}{}
		parts = append(parts, p)
	}

	if projectBin != "" {
		add(projectBin)
	}
	for _, p := range strings.Split(os.Getenv("PATH"), string(os.PathListSeparator)) {
		if isAllowlistedAmbientPath(p) {
			add(p)
		}
	}
	add("/usr/local/bin")
	add("/usr/bin")
	add("/bin")
	if runtime.GOOS == "darwin" {
		add("/opt/homebrew/bin")
	}
	return strings.Join(parts, string(os.PathListSeparator))
}

func isAllowlistedAmbientPath(p string) bool {
	switch {
	case strings.HasPrefix(p, "/opt/hostedtoolcache/"):
		return true
	case strings.HasPrefix(p, "/home/runner/go/"):
		return true
	case strings.HasPrefix(p, "/usr/local/go/"):
		return true
	default:
		return false
	}
}

// ApplyPinnedExecEnv sets cmd.Env with a pinned PATH. If cmd.Env is already set,
// non-PATH entries are preserved (e.g. DOCKER_IMAGE from e2e taskEnv); PATH is always replaced.
func ApplyPinnedExecEnv(cmd *exec.Cmd, projectBin string) {
	base := cmd.Env
	if base == nil {
		base = environForSubprocess()
	}
	pinned := PinnedExecPath(projectBin)
	out := make([]string, 0, len(base)+1)
	for _, e := range base {
		if strings.HasPrefix(e, "PATH=") {
			continue
		}
		out = append(out, e)
	}
	out = append(out, "PATH="+pinned)
	cmd.Env = out
}

func projectBinDir() string {
	dir, err := GetProjectDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "bin")
}
