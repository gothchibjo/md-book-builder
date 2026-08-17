package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config describes a single book build.
type Config struct {
	Title           string   `yaml:"title"`
	Subtitle        string   `yaml:"subtitle"`
	KBRoot          string   `yaml:"kb_root"`
	Out             string   `yaml:"out"`
	TOCLevels       []int    `yaml:"toc_levels"`
	TOC             *bool    `yaml:"toc,omitempty"`
	Cover           *bool    `yaml:"cover,omitempty"`
	Include         []string `yaml:"include"`
	Exclude         []string `yaml:"exclude,omitempty"`
	TransitiveLinks *bool    `yaml:"transitive_links,omitempty"`
	ChromePath      string   `yaml:"chrome_path,omitempty"`
	Locale          string   `yaml:"locale,omitempty"`
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

// Transitive reports whether link-follow collection reaches transitively
// through documents pulled in by links. It defaults to true when the option
// is omitted from the config.
func (c *Config) Transitive() bool {
	return c.TransitiveLinks == nil || *c.TransitiveLinks
}

// ShowTOC reports whether the table of contents should be rendered. It
// defaults to true when the option is omitted from the config.
func (c *Config) ShowTOC() bool {
	return c.TOC == nil || *c.TOC
}

// ShowCover reports whether the title page should be rendered. It defaults to
// true when the option is omitted from the config.
func (c *Config) ShowCover() bool {
	return c.Cover == nil || *c.Cover
}
