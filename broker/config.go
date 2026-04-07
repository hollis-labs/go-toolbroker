package broker

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed default-rules.yaml
var defaultRulesYAML []byte

// Config is the top-level tool broker configuration.
type Config struct {
	Rules []Rule `json:"rules" yaml:"rules"`
}

// LoadConfig reads a JSON config file and returns the parsed configuration.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// LoadRulesFromFile reads a rule file and auto-detects format by extension.
// Supported extensions: .yaml, .yml (YAML), .json (JSON).
// Returns an error for unknown extensions.
func LoadRulesFromFile(path string) ([]Rule, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read rules %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	var cfg Config

	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse YAML rules %s: %w", path, err)
		}
	case ".json":
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse JSON rules %s: %w", path, err)
		}
	default:
		return nil, fmt.Errorf("unsupported config format %q: use .yaml, .yml, or .json", ext)
	}

	return cfg.Rules, nil
}

// DefaultRules returns sensible defaults for Fragments Engine projects loaded from
// the embedded default-rules.yaml. These replicate the current mentat-chat
// behavior (exclude hadron_bp_*) and add intent-specific selections for
// common workflows.
func DefaultRules() []Rule {
	var cfg Config
	if err := yaml.Unmarshal(defaultRulesYAML, &cfg); err != nil {
		// Embedded file is validated at build time; this should never happen.
		panic(fmt.Sprintf("parse embedded default-rules.yaml: %v", err))
	}
	return cfg.Rules
}
