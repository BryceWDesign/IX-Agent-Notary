package receipt

import (
	"crypto/ed25519"
	"errors"
	"fmt"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/crypto"
)

type SignatureValidationOptions struct {
	Strict        bool
	PublicKeyPath string // optional override for public key lookup (base64url file)
}

type SignatureCheck struct {
	Skipped bool
	Alg     string
	KeyID   string
}

func ValidateSignature(r Receipt, opts SignatureValidationOptions) (*SignatureCheck, error) {
	integrity, ok := r["integrity"].(map[string]any)
	if !ok {
		return nil, errors.New("missing integrity object")
	}

	sigObj, ok := integrity["signature"].(map[string]any)
	if !ok {
		return nil, errors.New("missing integrity.signature object")
	}

	alg, _ := sigObj["alg"].(string)
	keyID, _ := sigObj["key_id"].(string)
	val, _ := sigObj["value"].(string)

	if isPlaceholder(val) {
		if opts.Strict {
			return nil, fmt.Errorf("signature is missing/placeholder")
		}
		return &SignatureCheck{Skipped: true, Alg: alg, KeyID: keyID}, nil
	}

	if alg != "ed25519" {
		return nil, fmt.Errorf("unsupported signature alg: %q (only ed25519 supported right now)", alg)
	}
	if keyID == "" {
		return nil, fmt.Errorf("missing integrity.signature.key_id")
	}

	var pub ed25519.PublicKey
	if opts.PublicKeyPath != "" {
		p, err := crypto.LoadEd25519PublicKeyFile(opts.PublicKeyPath)
		if err != nil {
			if opts.Strict {
				return nil, err
			}
			return &SignatureCheck{Skipped: true, Alg: alg, KeyID: keyID}, nil
		}
		pub = p
	} else {
		p, _, err := crypto.ResolvePublicKeyByID(keyID)
		if err != nil {
			if opts.Strict {
				return nil, err
			}
			return &SignatureCheck{Skipped: true, Alg: alg, KeyID: keyID}, nil
		}
		pub = p
	}

	sigBytes, err := crypto.DecodeBase64URLNoPad(val)
	if err != nil {
		return nil, fmt.Errorf("decode signature value: %w", err)
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return nil, fmt.Errorf("invalid ed25519 signature length: %d", len(sigBytes))
	}

	msg, err := canonicalBytesForSignature(r)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(pub, msg, sigBytes) {
		return nil, fmt.Errorf("signature verification failed (key_id=%s)", keyID)
	}

	return &SignatureCheck{Skipped: false, Alg: alg, KeyID: keyID}, nil
}

func canonicalBytesForSignature(r Receipt) ([]byte, error) {
	// Deep-clone via JSON round-trip to avoid mutating the original receipt.
	cloned, err := cloneReceipt(r)
	if err != nil {
		return nil, err
	}

	// Remove signature.value before canonicalization (normative rule in spec).
	if integrity, ok := cloned["integrity"].(map[string]any); ok {
		if sigObj, ok := integrity["signature"].(map[string]any); ok {
			delete(sigObj, "value")
		}
	}

	b, err := canon.CanonicalizeRFC8785(cloned)
	if err != nil {
		return nil, fmt.Errorf("canonicalize receipt for signature: %w", err)
	}
	return b, nil
}

func cloneReceipt(r Receipt) (map[string]any, error) {
	// Use the fact that Receipt is map[string]any; marshal/unmarshal gives a safe deep clone.
	// This is acceptable for verification tooling (not a hot path).
	b, err := marshalJSON(r)
	if err != nil {
		return nil, err
	}
	return unmarshalJSONObject(b)
}
