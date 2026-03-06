package receipt

import (
	"testing"
)

func TestGeneratedReceiptChain_Verifies(t *testing.T) {
	dir, leafPath, pubPath := buildChainedReceiptFixture(t)

	leaf, err := Load(leafPath)
	if err != nil {
		t.Fatalf("Load leaf: %v", err)
	}

	if _, err := ValidateCoreHashes(leaf, HashValidationOptions{Strict: true}); err != nil {
		t.Fatalf("ValidateCoreHashes leaf: %v", err)
	}
	if _, err := ValidateSignature(leaf, SignatureValidationOptions{
		Strict:        true,
		PublicKeyPath: pubPath,
	}); err != nil {
		t.Fatalf("ValidateSignature leaf: %v", err)
	}

	resolver, err := NewDirResolver(dir)
	if err != nil {
		t.Fatalf("NewDirResolver: %v", err)
	}

	validateParent := func(r Receipt) error {
		if _, err := ValidateCoreHashes(r, HashValidationOptions{Strict: true}); err != nil {
			return err
		}
		if _, err := ValidateSignature(r, SignatureValidationOptions{
			Strict:        true,
			PublicKeyPath: pubPath,
		}); err != nil {
			return err
		}
		return nil
	}

	cc, err := ValidateChain(leaf, resolver, validateParent, ChainValidationOptions{Strict: true})
	if err != nil {
		t.Fatalf("ValidateChain: %v", err)
	}

	if cc.Skipped {
		t.Fatalf("expected chain not skipped")
	}
	if cc.Depth != 1 {
		t.Fatalf("expected depth=1, got %d", cc.Depth)
	}
	if cc.RootReceiptID == "" {
		t.Fatalf("expected non-empty root receipt id")
	}
}
