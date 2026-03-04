package receipt

import (
	"strings"
	"testing"
)

func TestCanonicalizeApprovalForSigning_RemovesSignatureValue(t *testing.T) {
	approval := map[string]any{
		"approval_id": "appr-001",
		"type":        "ticket",
		"status":      "approved",
		"approver": map[string]any{
			"type":    "user",
			"id":      "you@example.com",
			"display": "You",
		},
		"scope": map[string]any{
			"kind":      "tool.invoke",
			"tool":      "filesystem",
			"operation": "file.write",
			"resource":  "docs/demo.txt",
		},
		"time": map[string]any{
			"requested_at": "2026-03-02T00:00:00Z",
			"decided_at":   "2026-03-02T00:00:01Z",
		},
		"notes": "demo",
		"signature": map[string]any{
			"alg":    "ed25519",
			"key_id": "approver-key-001",
			"value":  "BASE64URL_SIGNATURE_PLACEHOLDER",
		},
	}

	c1, err := CanonicalizeApprovalForSigning(approval)
	if err != nil {
		t.Fatalf("CanonicalizeApprovalForSigning: %v", err)
	}
	c2, err := CanonicalizeApprovalForSigning(approval)
	if err != nil {
		t.Fatalf("CanonicalizeApprovalForSigning (2): %v", err)
	}

	if string(c1) != string(c2) {
		t.Fatalf("expected stable canonicalization; got different outputs")
	}

	if strings.Contains(string(c1), "BASE64URL_SIGNATURE_PLACEHOLDER") {
		t.Fatalf("expected signature.value to be excluded from canonical bytes")
	}
	if strings.Contains(string(c1), "\"value\"") {
		t.Fatalf("expected signature.value field to be removed from canonical bytes")
	}
}
