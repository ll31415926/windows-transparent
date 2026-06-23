package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"windows-transparent/internal/opacity"
	"windows-transparent/internal/window"
)

type Rule struct {
	Process       string `json:"process"`
	Class         string `json:"class,omitempty"`
	TitleContains string `json:"title_contains,omitempty"`
	Opacity       int    `json:"opacity"`
	Enabled       *bool  `json:"enabled,omitempty"`
}

type Config struct {
	Rules []Rule `json:"rules"`
}

func DefaultPath() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("find user config directory: %w", err)
	}

	return filepath.Join(base, "wtrans", "config.json"), nil
}

func ResolvePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path != "" {
		return path, nil
	}

	return DefaultPath()
}

func Load(path string) (Config, error) {
	resolved, err := ResolvePath(path)
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, fmt.Errorf("config file not found at %s", resolved)
		}

		return Config{}, fmt.Errorf("open config %s: %w", resolved, err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg, err := Parse(file)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", resolved, err)
	}

	return cfg, nil
}

func LoadOrEmpty(path string) (Config, error) {
	resolved, err := ResolvePath(path)
	if err != nil {
		return Config{}, err
	}

	file, err := os.Open(resolved)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, nil
		}

		return Config{}, fmt.Errorf("open config %s: %w", resolved, err)
	}
	defer func() {
		_ = file.Close()
	}()

	cfg, err := Parse(file)
	if err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", resolved, err)
	}

	return cfg, nil
}

func Parse(r io.Reader) (Config, error) {
	var cfg Config
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return Config{}, fmt.Errorf("unexpected extra JSON value")
		}
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	resolved, err := ResolvePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	file, err := os.Create(resolved)
	if err != nil {
		return fmt.Errorf("create config %s: %w", resolved, err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("write config %s: %w", resolved, err)
	}

	return nil
}

func (c Config) Validate() error {
	for i, rule := range c.Rules {
		if strings.TrimSpace(rule.Process) == "" {
			return fmt.Errorf("rule %d process is required", i+1)
		}

		if err := opacity.Validate(rule.Opacity); err != nil {
			return fmt.Errorf("rule %d for %s: %w", i+1, rule.Process, err)
		}
	}

	return nil
}

func (c *Config) UpsertRule(rule Rule) bool {
	rule.Process = strings.TrimSpace(rule.Process)
	rule.Class = strings.TrimSpace(rule.Class)
	rule.TitleContains = strings.TrimSpace(rule.TitleContains)

	for i := range c.Rules {
		if c.Rules[i].SameSelector(rule) {
			c.Rules[i].Process = rule.Process
			c.Rules[i].Class = rule.Class
			c.Rules[i].TitleContains = rule.TitleContains
			c.Rules[i].Opacity = rule.Opacity
			c.Rules[i].Enabled = rule.Enabled
			return true
		}
	}

	c.Rules = append(c.Rules, rule)
	return false
}

func (r Rule) EnabledValue() bool {
	return r.Enabled == nil || *r.Enabled
}

func (r Rule) Matches(win window.Window) bool {
	if !r.EnabledValue() {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(r.Process), strings.TrimSpace(win.Process)) {
		return false
	}
	if strings.TrimSpace(r.Class) != "" && !strings.EqualFold(strings.TrimSpace(r.Class), strings.TrimSpace(win.Class)) {
		return false
	}

	titleContains := strings.TrimSpace(r.TitleContains)
	if titleContains != "" && !strings.Contains(strings.ToLower(win.Title), strings.ToLower(titleContains)) {
		return false
	}

	return true
}

func (r Rule) SameSelector(other Rule) bool {
	return strings.EqualFold(strings.TrimSpace(r.Process), strings.TrimSpace(other.Process)) &&
		strings.EqualFold(strings.TrimSpace(r.Class), strings.TrimSpace(other.Class)) &&
		strings.EqualFold(strings.TrimSpace(r.TitleContains), strings.TrimSpace(other.TitleContains))
}

func (r Rule) Describe() string {
	parts := []string{fmt.Sprintf("process=%q", strings.TrimSpace(r.Process))}
	if strings.TrimSpace(r.Class) != "" {
		parts = append(parts, fmt.Sprintf("class=%q", strings.TrimSpace(r.Class)))
	}
	if strings.TrimSpace(r.TitleContains) != "" {
		parts = append(parts, fmt.Sprintf("title_contains=%q", strings.TrimSpace(r.TitleContains)))
	}

	return strings.Join(parts, " ")
}

func Bool(value bool) *bool {
	return &value
}
