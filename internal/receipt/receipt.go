package receipt

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"ix-agent-notary/internal/canon"
	"ix-agent-notary/internal/hash"
)

type Receipt map[string]any

type HashValidationOptions struct {
	// If Strict is true, placeholder or missing hashes become errors.
	Strict bool
}

type HashCheck struct {
	Skipped bool

	ActionParametersExpected string
	ActionParametersComputed string

	ResultOutputExpected string
	ResultOutputComputed string
}

func Load(path string) (Receipt, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open receipt: %w", err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("read receipt: %w", err)
	}

	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("parse receipt json: %w", err)
	}

	m, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("receipt root must be a JSON object")
	}

	return Receipt(m), nil
}

// ComputeCoreHashes computes the canonical hashes for:
// - action.parameters
// - result.output
//
// Hash format convention in this repo:
// - "sha256:<digest>"
// - <digest> encoding is chosen by integrity.hash.encoding (hex|base64url)
func ComputeCoreHashes(r Receipt) (*HashCheck, error) {
	enc, err := receiptHashEncoding(r)
	if err != nil {
		return nil, err
	}

	paramsAny, paramsExpected, err := getActionParameters(r)
	if err != nil {
		return nil, err
	}
	outAny, outExpected, err := getResultOutput(r)
	if err != nil {
		return nil, err
	}

	paramsComputed, err := computeValueHash(paramsAny, enc)
	if err != nil {
		return nil, fmt.Errorf("compute action.parameters_hash: %w", err)
	}
	outComputed, err := computeValueHash(outAny, enc)
	if err != nil {
		return nil, fmt.Errorf("compute result.output_hash: %w", err)
	}

	return &HashCheck{
		ActionParametersExpected: paramsExpected,
		ActionParametersComputed: paramsComputed,
		ResultOutputExpected:     outExpected,
		ResultOutputComputed:     outComputed,
	}, nil
}

func ValidateCoreHashes(r Receipt, opts HashValidationOptions) (*HashCheck, error) {
	h, err := ComputeCoreHashes(r)
	if err != nil {
		return nil, err
	}

	paramsIsPlaceholder := isPlaceholder(h.ActionParametersExpected)
	outIsPlaceholder := isPlaceholder(h.ResultOutputExpected)

	if paramsIsPlaceholder || outIsPlaceholder {
		h.Skipped = true
		if opts.Strict {
			missing := []string{}
			if paramsIsPlaceholder {
				missing = append(missing, "action.parameters_hash")
			}
			if outIsPlaceholder {
				missing = append(missing, "result.output_hash")
			}
			return nil, fmt.Errorf("hash placeholders/missing not allowed in strict mode: %s", strings.Join(missing, ", "))
		}
		return h, nil
	}

	if !stringsEqualTrim(h.ActionParametersExpected, h.ActionParametersComputed) {
		return nil, fmt.Errorf("action.parameters_hash mismatch: expected %q computed %q", h.ActionParametersExpected, h.ActionParametersComputed)
	}
	if !stringsEqualTrim(h.ResultOutputExpected, h.ResultOutputComputed) {
		return nil, fmt.Errorf("result.output_hash mismatch: expected %q computed %q", h.ResultOutputExpected, h.ResultOutputComputed)
	}

	return h, nil
}

func computeValueHash(v any, enc hash.Encoding) (string, error) {
	canonBytes, err := canon.CanonicalizeRFC8785(v)
	if err != nil {
		return "", err
	}
	d := hash.Sha256Digest(canonBytes)
	ds, err := hash.EncodeDigest(d, enc)
	if err != nil {
		return "", err
	}
	return "sha256:" + ds, nil
}

func receiptHashEncoding(r Receipt) (hash.Encoding, error) {
	// Defaults align with the examples.
	alg := "sha-256"
	enc := "base64url"

	integrity, ok := r["integrity"].(map[string]any)
	if !ok {
		return "", errors.New("missing integrity object")
	}
	h, ok := integrity["hash"].(map[string]any)
	if !ok {
		return "", errors.New("missing integrity.hash object")
	}

	if v, ok := h["alg"].(string); ok && strings.TrimSpace(v) != "" {
		alg = v
	}
	a := strings.ToLower(strings.TrimSpace(alg))
	if a != "sha-256" && a != "sha256" {
		return "", fmt.Errorf("unsupported hash algorithm (only sha-256 supported right now): %q", alg)
	}

	if v, ok := h["encoding"].(string); ok && strings.TrimSpace(v) != "" {
		enc = v
	}

	parsed, err := hash.ParseEncoding(enc)
	if err != nil {
		return "", err
	}
	return parsed, nil
}

func getActionParameters(r Receipt) (any, string, error) {
	a, ok := r["action"].(map[string]any)
	if !ok {
		return nil, "", errors.New("missing action object")
	}
	params, ok := a["parameters"]
	if !ok {
		return nil, "", errors.New("missing action.parameters")
	}
	exp, _ := a["parameters_hash"].(string)
	return params, exp, nil
}

func getResultOutput(r Receipt) (any, string, error) {
	res, ok := r["result"].(map[string]any)
	if !ok {
		return nil, "", errors.New("missing result object")
	}
	out, ok := res["output"]
	if !ok {
		return nil, "", errors.New("missing result.output")
	}
	exp, _ := res["output_hash"].(string)
	return out, exp, nil
}

func isPlaceholder(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	return strings.Contains(strings.ToUpper(s), "PLACEHOLDER")
}

func stringsEqualTrim(a, b string) bool {
	return strings.TrimSpace(a) == strings.TrimSpace(b)
}
