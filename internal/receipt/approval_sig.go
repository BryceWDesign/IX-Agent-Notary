package receipt

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type ApprovalSigValidationOptions struct {
	// Strict fails if an approval exists but is missing a signature, or has an invalid signature.
	Strict bool

	// PublicKeyPath overrides key resolution and uses this pubkey for verification.
	// Expected format: base64url (no padding) encoded 32-byte ed25519 public key.
	PublicKeyPath string
}

type ApprovalSigCheck struct {
	Skipped  bool
	Total    int
	Verified int
}

func ValidateApprovalSignatures(r Receipt, opts ApprovalSigValidationOptions) (*ApprovalSigCheck, error) {
	pol, ok := r["policy"].(map[string]any)
	if !ok {
		// schema should prevent this; keep defensive
		if opts.Strict {
			return nil, errors.New("approval sig: missing policy object")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	apprsAny, ok := pol["approvals"]
	if !ok {
		if opts.Strict {
			return nil, errors.New("approval sig: missing policy.approvals")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	apprs, ok := apprsAny.([]any)
	if !ok {
		if opts.Strict {
			return nil, errors.New("approval sig: policy.approvals is not an array")
		}
		return &ApprovalSigCheck{Skipped: true}, nil
	}

	// No approvals is fine even in strict mode.
	if len(apprs) == 0 {
		return &ApprovalSigCheck{Skipped: false, Total: 0, Verified: 0}, nil
	}

	check := &ApprovalSigCheck{Skipped: false, Total: len(apprs), Verified: 0}

	// Optional override: one pubkey for all approvals (demo-friendly).
	var overridePub ed25519.PublicKey
	if strings.TrimSpace(opts.PublicKeyPath) != "" {
		pub, err := loadEd25519PublicKeyBase64URLFile(opts.PublicKeyPath)
		if err != nil {
			return nil, err
		}
		overridePub = pub
	}

	for i, a := range apprs {
		obj, ok := a.(map[string]any)
		if !ok {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d] is not an object", i)
			}
			continue
		}

		sigObjAny, hasSig := obj["signature"]
		if !hasSig || sigObjAny == nil {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d] missing signature", i)
			}
			continue
		}

		sigObj, ok := sigObjAny.(map[string]any)
		if !ok {
			if opts.Strict {
				return nil, fmt.Errorf("approval sig: approvals[%d].signature is not an object", i)
			}
			continue
		}

		alg, _ := sigObj["alg"].(string)
		keyID, _ := sigObj["key_id"].(string)
		val, _ := sigObj["value"].(string)

		alg = strings.ToLower(strings.TrimSpace(alg))
		keyID = strings.TrimSpace(keyID)
		val = strings.TrimSpace(val)

		if alg != "ed25519" {
			return nil, fmt.Errorf("approval sig: approvals[%d] unsupported alg %q", i, alg)
		}
		if keyID == "" {
			return nil, fmt.Errorf("approval sig: approvals[%d] missing signature.key_id", i)
		}
		if val == "" {
			return nil, fmt.Errorf("approval sig: approvals[%d] missing signature.value", i)
		}

		payload, err := CanonicalizeApprovalForSigning(obj)
		if err != nil {
			return nil, fmt.Errorf("approval sig: approvals[%d] canonicalize: %w", i, err)
		}

		sigBytes, err := base64.RawURLEncoding.DecodeString(val)
		if err != nil {
			return nil, fmt.Errorf("approval sig: approvals[%d] signature.value not base64url: %w", i, err)
		}
		if len(sigBytes) != ed25519.SignatureSize {
			return nil, fmt.Errorf("approval sig: approvals[%d] signature size invalid (got %d)", i, len(sigBytes))
		}

		var pub ed25519.PublicKey
		if overridePub != nil {
			pub = overridePub
		} else {
			pub, err = resolveEd25519PublicKeyByKeyID(keyID)
			if err != nil {
				return nil, fmt.Errorf("approval sig: approvals[%d] resolve pubkey: %w", i, err)
			}
		}

		if !ed25519.Verify(pub, payload, sigBytes) {
			return nil, fmt.Errorf("approval sig: approvals[%d] invalid signature", i)
		}

		check.Verified++
	}

	// If strict, require that every approval is signed+valid.
	if opts.Strict && check.Verified != check.Total {
		return nil, fmt.Errorf("approval sig: strict mode requires all approvals be signed (%d/%d verified)", check.Verified, check.Total)
	}

	return check, nil
}

func resolveEd25519PublicKeyByKeyID(keyID string) (ed25519.PublicKey, error) {
	root, err := repoRootForReceiptPackage()
	if err != nil {
		return nil, err
	}
	p := filepath.Join(root, "keys", "dev", keyID+".pub")
	return loadEd25519PublicKeyBase64URLFile(p)
}

func loadEd25519PublicKeyBase64URLFile(path string) (ed25519.PublicKey, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("approval sig: read pubkey: %w", err)
	}

	raw, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(string(b)))
	if err != nil {
		return nil, fmt.Errorf("approval sig: decode pubkey (base64url): %w", err)
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("approval sig: pubkey must be %d bytes (got %d)", ed25519.PublicKeySize, len(raw))
	}
	return ed25519.PublicKey(raw), nil
}

func repoRootForReceiptPackage() (string, error) {
	// This file is: <root>/internal/receipt/approval_sig.go
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errors.New("approval sig: runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..")), nil
}
