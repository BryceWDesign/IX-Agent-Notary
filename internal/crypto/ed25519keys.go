package crypto

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func LoadEd25519PrivateKeyFromSeedFile(path string) (ed25519.PrivateKey, error) {
	b, err := readB64URLFile(path)
	if err != nil {
		return nil, err
	}

	switch len(b) {
	case ed25519.SeedSize:
		return ed25519.NewKeyFromSeed(b), nil
	case ed25519.PrivateKeySize:
		return ed25519.PrivateKey(b), nil
	default:
		return nil, fmt.Errorf("unsupported ed25519 private key length: %d (expected 32 seed bytes or 64 private key bytes)", len(b))
	}
}

func LoadEd25519PublicKeyFile(path string) (ed25519.PublicKey, error) {
	b, err := readB64URLFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("unsupported ed25519 public key length: %d (expected 32 bytes)", len(b))
	}
	return ed25519.PublicKey(b), nil
}

// ResolvePublicKeyByID searches for:
// - keys/<keyID>.pub
// - keys/dev/<keyID>.pub
func ResolvePublicKeyByID(keyID string) (ed25519.PublicKey, string, error) {
	keyID = strings.TrimSpace(keyID)
	if keyID == "" {
		return nil, "", fmt.Errorf("key_id is empty")
	}

	candidates := []string{
		filepath.Join("keys", keyID+".pub"),
		filepath.Join("keys", "dev", keyID+".pub"),
	}

	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			pub, err := LoadEd25519PublicKeyFile(p)
			if err != nil {
				return nil, "", err
			}
			return pub, p, nil
		}
	}

	return nil, "", fmt.Errorf("could not resolve public key for key_id %q (looked in keys/ and keys/dev/)", keyID)
}

func readB64URLFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}
	s := strings.TrimSpace(string(raw))
	if s == "" {
		return nil, fmt.Errorf("key file is empty: %s", path)
	}
	b, err := DecodeBase64URLNoPad(s)
	if err != nil {
		return nil, fmt.Errorf("decode key file %s: %w", path, err)
	}
	return b, nil
}
