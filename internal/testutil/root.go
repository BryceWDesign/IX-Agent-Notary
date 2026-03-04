package testutil

import (
	"path/filepath"
	"runtime"
	"testing"
)

func RepoRoot(t *testing.T) string {
	t.Helper()

	// This file lives at: <root>/internal/testutil/root.go
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}
