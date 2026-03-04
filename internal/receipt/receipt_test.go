package receipt

import (
	"path/filepath"
	"testing"

	"ix-agent-notary/internal/testutil"
)

func TestExamples_StrictHashesPass(t *testing.T) {
	root := testutil.RepoRoot(t)

	cases := []struct {
		name string
		path string
	}{
		{"minimal", filepath.Join(root, "examples", "receipts", "minimal.receipt.json")},
		{"denied", filepath.Join(root, "examples", "receipts", "denied.receipt.json")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Load(tc.path)
			if err != nil {
				t.Fatalf("Load: %v", err)
			}

			if _, err := ValidateCoreHashes(r, HashValidationOptions{Strict: true}); err != nil {
				t.Fatalf("ValidateCoreHashes (strict): %v", err)
			}
		})
	}
}

func TestExamples_ExpectedComputedHashesMatch(t *testing.T) {
	root := testutil.RepoRoot(t)

	r, err := Load(filepath.Join(root, "examples", "receipts", "minimal.receipt.json"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	h, err := ComputeCoreHashes(r)
	if err != nil {
		t.Fatalf("ComputeCoreHashes: %v", err)
	}

	if h.ActionParametersExpected != h.ActionParametersComputed {
		t.Fatalf("parameters_hash mismatch expected=%q computed=%q", h.ActionParametersExpected, h.ActionParametersComputed)
	}
	if h.ResultOutputExpected != h.ResultOutputComputed {
		t.Fatalf("output_hash mismatch expected=%q computed=%q", h.ResultOutputExpected, h.ResultOutputComputed)
	}
}
