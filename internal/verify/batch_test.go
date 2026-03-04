package verify

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestVerifyDir_Examples_StrictChain(t *testing.T) {
	root := testutil.RepoRoot(t)

	_, err := VerifyDir(DirOptions{
		Dir:             filepath.Join(root, "examples", "receipts"),
		SchemaPath:      filepath.Join(root, "spec", "receipt.schema.json"),
		StrictHashes:    true,
		StrictSignature: true,
		StrictChain:     true,
	})
	if err != nil {
		t.Fatalf("VerifyDir failed: %v", err)
	}
}
