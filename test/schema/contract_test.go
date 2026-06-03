package schema

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestCRDSpecOpenAPIFragments(t *testing.T) {
	root := repoRoot(t)
	update := os.Getenv("UPDATE_GOLDEN") == "1"

	for _, tc := range DefaultCases {
		t.Run(tc.GoldenFile, func(t *testing.T) {
			t.Parallel()
			got, err := ExtractSpecOpenAPIFragment(CRDPath(root, tc.CRDFile))
			if err != nil {
				t.Fatalf("extract: %v", err)
			}

			goldenPath := GoldenPath(root, tc.GoldenFile)
			if update {
				if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("mkdir golden dir: %v", err)
				}
				if err := os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatalf("write golden: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("read golden %q: %v (run UPDATE_GOLDEN=1 go test ./test/schema/ -run TestCRDSpecOpenAPIFragments)", goldenPath, err)
			}
			if !bytes.Equal(got, want) {
				t.Fatalf(
					"OpenAPI spec fragment drift for %s\nrun: UPDATE_GOLDEN=1 go test ./test/schema/ -run TestCRDSpecOpenAPIFragments\nor: task test:schema:update",
					tc.CRDFile,
				)
			}
		})
	}
}
