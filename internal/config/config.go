package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config describes a single book build.
type Config struct {
	Title      string   `yaml:"title"`
	Subtitle   string   `yaml:"subtitle"`
	KBRoot     string   `yaml:"kb_root"`
	Out        string   `yaml:"out"`
	TOCLevels  []int    `yaml:"toc_levels"`
	Include    []string `yaml:"include"`
	ChromePath string   `yaml:"chrome_path,omitempty"`
}

// Load reads and validates a Config from path. All relative paths are
// resolved against the config file directory. TOCLevels default to [1]
// (documents only) when omitted.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}

	if len(c.TOCLevels) == 0 {
		c.TOCLevels = []int{1}
	}
	if c.Title == "" {
		return nil, fmt.Errorf("config %s: title is required", path)
	}
	if c.KBRoot == "" {
		return nil, fmt.Errorf("config %s: kb_root is required", path)
	}
	if c.Out == "" {
		return nil, fmt.Errorf("config %s: out is required", path)
	}
	if len(c.Include) == 0 {
		return nil, fmt.Errorf("config %s: include must contain at least one pattern", path)
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	base := filepath.Dir(abs)

	c.KBRoot = resolveRelative(base, c.KBRoot)
	c.Out = resolveRelative(base, c.Out)
	c.ChromePath = resolveRelative(base, c.ChromePath)
	return &c, nil
}

func resolveRelative(base, p string) string {
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}
