package receipt

import (
	"crypto/ed25519"
	"path/filepath"
	"testing"
	"time"

	"ix-agent-notary/internal/crypto"
	"ix-agent-notary/internal/id"
	"ix-agent-notary/internal/policy"
	"ix-agent-notary/internal/testutil"
)

const receiptTestKeyID = "test-key-001"

func newSignedTestReceipt(t *testing.T, targetPath string, seedPath string, keyID string) Receipt {
	t.Helper()

	root := testutil.RepoRoot(t)

	p, err := policy.Load(filepath.Join(root, "policy", "demo.policy.json"))
	if err != nil {
		t.Fatalf("policy.Load: %v", err)
	}

	dec := p.Evaluate(policy.Request{
		Kind:      "tool.invoke",
		Tool:      "filesystem",
		Operation: "file.write",
		Path:      targetPath,
	})

	now := time.Now().UTC().Format(time.RFC3339)

	receiptID, err := id.NewUUIDv4()
	if err != nil {
		t.Fatalf("id.NewUUIDv4 receipt_id: %v", err)
	}
	traceID, err := id.NewUUIDv4()
	if err != nil {
		t.Fatalf("id.NewUUIDv4 trace_id: %v", err)
	}

	status := "denied"
	summary := "Denied filesystem file.write (simulated)."
	output := map[string]any{
		"written": false,
		"denied":  true,
	}

	if dec.Decision == "allow" {
		status = "success"
		summary = "Allowed filesystem file.write (simulated)."
		output = map[string]any{
			"path":             targetPath,
			"written":          true,
			"content_redacted": true,
		}
	}

	rules := []map[string]any{}
	for _, mr := range dec.Matched {
		rules = append(rules, map[string]any{
			"rule_id":     mr.RuleID,
			"effect":      mr.Effect,
			"explanation": mr.Explanation,
		})
	}

	r := Receipt{
		"receipt_version": "0.1.0",
		"receipt_id":      receiptID,
		"time": map[string]any{
			"requested_at": now,
			"decided_at":   now,
			"completed_at": now,
		},
		"trace": map[string]any{
			"trace_id": traceID,
			"step":     1,
		},
		"actor": map[string]any{
			"type":       "agent",
			"id":         "agent:test",
			"display":    "Test Agent",
			"session_id": "sess-test-001",
		},
		"notary": map[string]any{
			"runtime":     "IX-Agent-Notary",
			"version":     "0.1.0-dev",
			"instance_id": "notary-test-001",
			"environment": "local",
		},
		"action": map[string]any{
			"kind":      "tool.invoke",
			"tool":      "filesystem",
			"operation": "file.write",
			"parameters": map[string]any{
				"path":             targetPath,
				"bytes":            10,
				"content_redacted": true,
			},
			"parameters_hash": "sha256:PLACEHOLDER_PARAMETERS_HASH",
		},
		"policy": map[string]any{
			"policy_id":     dec.PolicyID,
			"policy_hash":   dec.PolicyHash,
			"policy_source": dec.PolicySource,
			"decision":      dec.Decision,
			"reason":        dec.Reason,
			"rules":         rules,
			"approvals":     []any{},
		},
		"result": map[string]any{
			"status":      status,
			"summary":     summary,
			"output":      output,
			"output_hash": "sha256:PLACEHOLDER_OUTPUT_HASH",
		},
		"integrity": map[string]any{
			"canonicalization": "RFC8785-JCS",
			"hash": map[string]any{
				"alg":      "sha-256",
				"encoding": "base64url",
			},
			"signature": map[string]any{
				"alg":    "ed25519",
				"key_id": keyID,
				"value":  "BASE64URL_SIGNATURE_PLACEHOLDER",
			},
		},
	}

	signReceiptForTest(t, r, seedPath, keyID)
	return r
}

func writeSignedTestReceipt(t *testing.T, path string, targetPath string, seedPath string, keyID string) Receipt {
	t.Helper()

	r := newSignedTestReceipt(t, targetPath, seedPath, keyID)
	if err := Write(path, r); err != nil {
		t.Fatalf("Write receipt %s: %v", path, err)
	}
	return r
}

func buildChainedReceiptFixture(t *testing.T) (dir string, leafPath string, pubPath string) {
	t.Helper()

	seedPath, pubPath := testutil.TempEd25519Keypair(t, receiptTestKeyID)
	dir = t.TempDir()

	parentPath := filepath.Join(dir, "parent.receipt.json")
	childPath := filepath.Join(dir, "child.receipt.json")

	parent := writeSignedTestReceipt(t, parentPath, "docs/parent.txt", seedPath, receiptTestKeyID)
	child := writeSignedTestReceipt(t, childPath, "docs/child.txt", seedPath, receiptTestKeyID)

	parentID := mustStringField(t, parent, "receipt_id")
	parentTrace := mustObjectField(t, parent, "trace")
	parentTraceID := mustStringField(t, parentTrace, "trace_id")

	childTrace := mustObjectField(t, child, "trace")
	childTrace["trace_id"] = parentTraceID
	childTrace["step"] = 2
	childTrace["parent_receipt_id"] = parentID

	signReceiptForTest(t, child, seedPath, receiptTestKeyID)

	if err := Write(childPath, child); err != nil {
		t.Fatalf("Write child receipt: %v", err)
	}

	return dir, childPath, pubPath
}

func signReceiptForTest(t *testing.T, r Receipt, seedPath string, keyID string) {
	t.Helper()

	hc, err := ComputeCoreHashes(r)
	if err != nil {
		t.Fatalf("ComputeCoreHashes: %v", err)
	}

	action := mustObjectField(t, r, "action")
	action["parameters_hash"] = hc.ActionParametersComputed

	result := mustObjectField(t, r, "result")
	result["output_hash"] = hc.ResultOutputComputed

	integrity := mustObjectField(t, r, "integrity")
	sigObj := mustObjectField(t, integrity, "signature")
	sigObj["alg"] = "ed25519"
	sigObj["key_id"] = keyID

	priv, err := crypto.LoadEd25519PrivateKeyFromSeedFile(seedPath)
	if err != nil {
		t.Fatalf("LoadEd25519PrivateKeyFromSeedFile: %v", err)
	}

	msg, err := canonicalBytesForSignature(r)
	if err != nil {
		t.Fatalf("canonicalBytesForSignature: %v", err)
	}

	sig := ed25519.Sign(priv, msg)
	sigObj["value"] = crypto.EncodeBase64URLNoPad(sig)
}

func mustObjectField(t *testing.T, obj map[string]any, key string) map[string]any {
	t.Helper()

	v, ok := obj[key].(map[string]any)
	if !ok || v == nil {
		t.Fatalf("field %q is missing or not an object", key)
	}
	return v
}

func mustStringField(t *testing.T, obj map[string]any, key string) string {
	t.Helper()

	v, ok := obj[key].(string)
	if !ok || v == "" {
		t.Fatalf("field %q is missing or not a non-empty string", key)
	}
	return v
}
