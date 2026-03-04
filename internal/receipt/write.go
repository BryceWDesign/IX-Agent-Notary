package receipt

import (
	"encoding/json"
	"fmt"
	"os"
)

func Write(path string, r Receipt) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal receipt: %w", err)
	}
	b = append(b, byte('\n'))
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("write receipt: %w", err)
	}
	return nil
}
