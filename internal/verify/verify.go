package verify

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"ix-agent-notary/internal/receipt"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Options struct {
	ReceiptPath  string
	SchemaPath   string
	StrictHashes bool
}

type Result struct {
	ReceiptPath string
	SchemaPath  string
	Hashes      receipt.HashCheck
}

func Run(opts Options) (*Result, error) {
	if opts.ReceiptPath == "" {
		return nil, errors.New("receipt path is required")
	}
	if opts.SchemaPath == "" {
		opts.SchemaPath = filepath.Join("spec", "receipt.schema.json")
	}

	schema, err := compileSchema(opts.SchemaPath)
	if err != nil {
		return nil, err
	}

	r, err := receipt.Load(opts.ReceiptPath)
	if err != nil {
		return nil, err
	}

	if err := schema.Validate(any(r)); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	hc, err := receipt.ValidateCoreHashes(r, receipt.HashValidationOptions{Strict: opts.StrictHashes})
	if err != nil {
		return nil, err
	}

	// NOTE: cryptographic signature verification is intentionally not implemented in this commit.
	// Upcoming commits will:
	// - verify integrity.signature.value against integrity.signature.key_id
	// - validate receipt chaining (parent_receipt_id) and optional artifact hashes

	return &Result{ReceiptPath: opts.ReceiptPath, SchemaPath: opts.SchemaPath, Hashes: *hc}, nil
}

func compileSchema(schemaPath string) (*jsonschema.Schema, error) {
	abs, err := filepath.Abs(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("resolve schema path: %w", err)
	}

	f, err := os.Open(abs)
	if err != nil {
		return nil, fmt.Errorf("open schema: %w", err)
	}
	defer f.Close()

	c := jsonschema.NewCompiler()

	// Use a stable internal URL for compilation. The schema content is loaded from disk.
	const schemaURL = "https://ix-agent-notary.local/spec/receipt.schema.json"

	if err := c.AddResource(schemaURL, f); err != nil {
		return nil, fmt.Errorf("add schema resource: %w", err)
	}

	s, err := c.Compile(schemaURL)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}

	return s, nil
}
