package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// PinnedExecPath returns the allowlisted PATH for test subprocesses (Sonar go:S4036).
// Base: /usr/bin:/bin:/usr/local/bin — CI installs kubectl/helm/kind/task under /usr/local/bin.
// Optional projectBin (repo bin/) for tools:install artifacts.
// On darwin, /opt/homebrew/bin is allowlisted so local Apple Silicon e2e finds brew tools.
func PinnedExecPath(projectBin string) string {
	parts := []string{"/usr/bin", "/bin", "/usr/local/bin"}
	if projectBin != "" {
		parts = append(parts, projectBin)
	}
	if runtime.GOOS == "darwin" {
		parts = append(parts, "/opt/homebrew/bin")
	}
	return strings.Join(parts, string(os.PathListSeparator))
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
