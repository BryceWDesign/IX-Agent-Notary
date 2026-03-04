package main

import (
	"flag"
	"fmt"
	"os"

	"ix-agent-notary/internal/receipt"
	"ix-agent-notary/internal/verify"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "verify":
		verifyCmd(os.Args[2:])
	case "hash":
		hashCmd(os.Args[2:])
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func verifyCmd(args []string) {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	schemaPath := fs.String("schema", "", "path to receipt JSON Schema (default: spec/receipt.schema.json)")
	strictHashes := fs.Bool("strict-hashes", false, "fail if parameters_hash/output_hash are placeholders or missing")

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "verify requires exactly 1 argument: <receipt.json>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an verify examples/receipts/minimal.receipt.json")
		os.Exit(2)
	}

	receiptPath := fs.Arg(0)

	res, err := verify.Run(verify.Options{
		ReceiptPath:  receiptPath,
		SchemaPath:   *schemaPath,
		StrictHashes: *strictHashes,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	// Keep output CI-friendly and honest about what's not implemented yet.
	if res.Hashes.Skipped {
		fmt.Println("OK: receipt matches schema; hash checks skipped (placeholders present); signature verification not yet implemented")
		return
	}

	fmt.Println("OK: receipt matches schema; hash fields match canonical SHA-256; signature verification not yet implemented")
}

func hashCmd(args []string) {
	fs := flag.NewFlagSet("hash", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if fs.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "hash requires exactly 1 argument: <receipt.json>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  ix-an hash examples/receipts/minimal.receipt.json")
		os.Exit(2)
	}

	receiptPath := fs.Arg(0)

	r, err := receipt.Load(receiptPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	h, err := receipt.ComputeCoreHashes(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("action.parameters_hash=", h.ActionParametersComputed)
	fmt.Println("result.output_hash=", h.ResultOutputComputed)
}

func usage() {
	fmt.Fprintln(os.Stderr, "IX-Agent-Notary (ix-an)")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  ix-an <command> [options]")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  verify   Validate a receipt against the JSON Schema (+ optional hash checks)")
	fmt.Fprintln(os.Stderr, "  hash     Compute canonical hashes for action.parameters and result.output")
	fmt.Fprintln(os.Stderr, "  help     Show this help")
	fmt.Fprintln(os.Stderr)
}
