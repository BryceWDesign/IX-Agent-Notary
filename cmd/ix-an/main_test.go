package main

import (
	"flag"
	"reflect"
	"testing"
)

func TestNormalizeInterspersedFlagsVerifyMovesFlagsAheadOfReceiptPath(t *testing.T) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	var strictChain bool
	var strictApprovals bool
	var chainDir string
	fs.BoolVar(&strictChain, "strict-chain", false, "")
	fs.StringVar(&chainDir, "chain-dir", "", "")
	fs.BoolVar(&strictApprovals, "strict-approvals", false, "")

	args := []string{
		"examples/receipts/denied.receipt.json",
		"--strict-chain",
		"--chain-dir",
		"examples/receipts",
		"--strict-approvals",
	}

	normalized, err := normalizeInterspersedFlags(fs, args)
	if err != nil {
		t.Fatalf("normalizeInterspersedFlags returned error: %v", err)
	}

	want := []string{
		"--strict-chain",
		"--chain-dir",
		"examples/receipts",
		"--strict-approvals",
		"examples/receipts/denied.receipt.json",
	}
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("normalized args mismatch\n got: %#v\nwant: %#v", normalized, want)
	}

	if err := fs.Parse(normalized); err != nil {
		t.Fatalf("Parse failed after normalization: %v", err)
	}
	if !strictChain {
		t.Fatal("strict-chain should be true after parsing normalized args")
	}
	if !strictApprovals {
		t.Fatal("strict-approvals should be true after parsing normalized args")
	}
	if chainDir != "examples/receipts" {
		t.Fatalf("chain-dir mismatch: got %q want %q", chainDir, "examples/receipts")
	}
	if got := fs.Arg(0); got != "examples/receipts/denied.receipt.json" {
		t.Fatalf("receipt path mismatch: got %q want %q", got, "examples/receipts/denied.receipt.json")
	}
}

func TestNormalizeInterspersedFlagsVerifyDirHandlesTrailingBoolFlag(t *testing.T) {
	fs := flag.NewFlagSet("verify-dir", flag.ContinueOnError)
	var strictApprovals bool
	var strictChain bool
	fs.BoolVar(&strictApprovals, "strict-approvals", false, "")
	fs.BoolVar(&strictChain, "strict-chain", true, "")

	args := []string{"examples/receipts", "--strict-approvals"}

	normalized, err := normalizeInterspersedFlags(fs, args)
	if err != nil {
		t.Fatalf("normalizeInterspersedFlags returned error: %v", err)
	}

	want := []string{"--strict-approvals", "examples/receipts"}
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("normalized args mismatch\n got: %#v\nwant: %#v", normalized, want)
	}

	if err := fs.Parse(normalized); err != nil {
		t.Fatalf("Parse failed after normalization: %v", err)
	}
	if !strictApprovals {
		t.Fatal("strict-approvals should be true after parsing normalized args")
	}
	if !strictChain {
		t.Fatal("strict-chain default should remain true")
	}
	if got := fs.Arg(0); got != "examples/receipts" {
		t.Fatalf("directory argument mismatch: got %q want %q", got, "examples/receipts")
	}
}

func TestNormalizeInterspersedFlagsSupportsEqualsSyntax(t *testing.T) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	var schema string
	var strictSignature bool
	fs.StringVar(&schema, "schema", "", "")
	fs.BoolVar(&strictSignature, "strict-signature", false, "")

	args := []string{
		"examples/receipts/minimal.receipt.json",
		"--schema=spec/receipt.schema.json",
		"--strict-signature",
	}

	normalized, err := normalizeInterspersedFlags(fs, args)
	if err != nil {
		t.Fatalf("normalizeInterspersedFlags returned error: %v", err)
	}

	want := []string{
		"--schema=spec/receipt.schema.json",
		"--strict-signature",
		"examples/receipts/minimal.receipt.json",
	}
	if !reflect.DeepEqual(normalized, want) {
		t.Fatalf("normalized args mismatch\n got: %#v\nwant: %#v", normalized, want)
	}

	if err := fs.Parse(normalized); err != nil {
		t.Fatalf("Parse failed after normalization: %v", err)
	}
	if schema != "spec/receipt.schema.json" {
		t.Fatalf("schema mismatch: got %q want %q", schema, "spec/receipt.schema.json")
	}
	if !strictSignature {
		t.Fatal("strict-signature should be true after parsing normalized args")
	}
	if got := fs.Arg(0); got != "examples/receipts/minimal.receipt.json" {
		t.Fatalf("receipt path mismatch: got %q want %q", got, "examples/receipts/minimal.receipt.json")
	}
}

func TestNormalizeInterspersedFlagsErrorsWhenValueIsMissing(t *testing.T) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	var chainDir string
	fs.StringVar(&chainDir, "chain-dir", "", "")

	_, err := normalizeInterspersedFlags(fs, []string{"receipt.json", "--chain-dir"})
	if err == nil {
		t.Fatal("expected error for missing flag value, got nil")
	}
}
