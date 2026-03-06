package crypto

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ResolvePublicKeyOptions struct {
	KeyID         string
	PublicKeyPath string
	SearchDirs    []string
}

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

func ResolveEd25519PublicKey(opts ResolvePublicKeyOptions) (ed25519.PublicKey, string, error) {
	if path := strings.TrimSpace(opts.PublicKeyPath); path != "" {
		pub, err := LoadEd25519PublicKeyFile(path)
		if err != nil {
			return nil, "", err
		}
		return pub, path, nil
	}

	keyID := strings.TrimSpace(opts.KeyID)
	if keyID == "" {
		return nil, "", fmt.Errorf("key_id is empty")
	}

	candidates := publicKeyCandidates(keyID, opts.SearchDirs)
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			pub, err := LoadEd25519PublicKeyFile(p)
			if err != nil {
				return nil, "", err
			}
			return pub, p, nil
		}
	}

	return nil, "", fmt.Errorf("could not resolve public key for key_id %q (looked in %s)", keyID, strings.Join(candidates, ", "))
}

func ResolvePublicKeyByID(keyID string) (ed25519.PublicKey, string, error) {
	return ResolveEd25519PublicKey(ResolvePublicKeyOptions{KeyID: keyID})
}

func DefaultPublicKeySearchDirs() []string {
	return []string{
		filepath.Join("keys"),
		filepath.Join("keys", "dev"),
	}
}

func publicKeyCandidates(keyID string, searchDirs []string) []string {
	dirs := normalizeSearchDirs(searchDirs)
	out := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		out = append(out, filepath.Join(dir, keyID+".pub"))
	}
	return out
}

func normalizeSearchDirs(searchDirs []string) []string {
	if len(searchDirs) == 0 {
		return DefaultPublicKeySearchDirs()
	}

	seen := map[string]struct{}{}
	out := make([]string, 0, len(searchDirs))

	for _, dir := range searchDirs {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		dir = filepath.Clean(dir)
		if _, ok := seen[dir]; ok {
			continue
		}
		seen[dir] = struct{}{}
		out = append(out, dir)
	}

	if len(out) == 0 {
		return DefaultPublicKeySearchDirs()
	}

	return out
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
