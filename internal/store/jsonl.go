package store

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"ix-agent-notary/internal/receipt"
)

func AppendJSONL(logPath string, r receipt.Receipt) error {
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	b, err := json.Marshal(r) // compact (good for append-only logs)
	if err != nil {
		return fmt.Errorf("marshal receipt: %w", err)
	}

	if _, err := f.Write(append(b, '\n')); err != nil {
		return fmt.Errorf("append log: %w", err)
	}
	return nil
}

func ReadAllJSONL(logPath string) ([]receipt.Receipt, error) {
	f, err := os.Open(logPath)
	if err != nil {
		return nil, fmt.Errorf("open log: %w", err)
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	// Bump buffer for larger receipts.
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)

	var out []receipt.Receipt
	lineNo := 0

	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}

		var v any
		if err := json.Unmarshal([]byte(line), &v); err != nil {
			return nil, fmt.Errorf("log line %d: invalid json: %w", lineNo, err)
		}

		m, ok := v.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("log line %d: expected json object", lineNo)
		}
		out = append(out, receipt.Receipt(m))
	}

	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan log: %w", err)
	}

	return out, nil
}
