package receipt

import (
	"errors"
	"fmt"
	"strings"
)

type MapResolver struct {
	byID map[string]Receipt
}

func NewMapResolver(byID map[string]Receipt) (*MapResolver, error) {
	if byID == nil {
		return nil, errors.New("map resolver: byID is nil")
	}
	return &MapResolver{byID: byID}, nil
}

func (mr *MapResolver) Resolve(receiptID string) (Receipt, string, error) {
	receiptID = strings.TrimSpace(receiptID)
	if receiptID == "" {
		return nil, "", errors.New("resolve: receiptID is empty")
	}

	r, ok := mr.byID[receiptID]
	if !ok {
		return nil, "", fmt.Errorf("resolve: receipt_id %q not found in log", receiptID)
	}
	return r, "memory", nil
}
