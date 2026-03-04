package policy

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type Policy struct {
	PolicyID      string `json:"policy_id"`
	DefaultEffect string `json:"default_effect"` // "allow" or "deny"
	Rules         []Rule `json:"rules"`
}

type Rule struct {
	RuleID      string `json:"rule_id"`
	Effect      string `json:"effect"` // "allow" or "deny"
	Kind        string `json:"kind,omitempty"`
	Tool        string `json:"tool,omitempty"`
	Operation   string `json:"operation,omitempty"`
	PathPrefix  string `json:"path_prefix,omitempty"`
	PathExact   string `json:"path_exact,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

func Load(path string) (*Policy, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read policy: %w", err)
	}
	var p Policy
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("parse policy json: %w", err)
	}
	if strings.TrimSpace(p.PolicyID) == "" {
		return nil, fmt.Errorf("policy_id is required")
	}
	if p.DefaultEffect == "" {
		p.DefaultEffect = "deny"
	}
	p.DefaultEffect = strings.ToLower(strings.TrimSpace(p.DefaultEffect))
	if p.DefaultEffect != "allow" && p.DefaultEffect != "deny" {
		return nil, fmt.Errorf("default_effect must be allow|deny (got %q)", p.DefaultEffect)
	}
	for i := range p.Rules {
		p.Rules[i].Effect = strings.ToLower(strings.TrimSpace(p.Rules[i].Effect))
		if p.Rules[i].Effect != "allow" && p.Rules[i].Effect != "deny" {
			return nil, fmt.Errorf("rule %d effect must be allow|deny (got %q)", i, p.Rules[i].Effect)
		}
	}
	return &p, nil
}
