package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"windows-transparent/internal/opacity"
)

type Rule struct {
	Process string `json:"process"`
	Opacity int    `json:"opacity"`
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
