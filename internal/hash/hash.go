package hash

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
)

type Encoding string

const (
	EncodingHex       Encoding = "hex"
	EncodingBase64URL Encoding = "base64url"
)

func ParseEncoding(s string) (Encoding, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "hex":
		return EncodingHex, nil
	case "base64url":
		return EncodingBase64URL, nil
	default:
		return "", fmt.Errorf("unsupported hash encoding: %q", s)
	}
}

func Sha256Digest(data []byte) [32]byte {
	return sha256.Sum256(data)
}

func EncodeDigest(d [32]byte, enc Encoding) (string, error) {
	switch enc {
	case EncodingHex:
		return hex.EncodeToString(d[:]), nil
	case EncodingBase64URL:
		return base64.RawURLEncoding.EncodeToString(d[:]), nil
	default:
		return "", fmt.Errorf("unsupported encoding: %q", enc)
	}
}
