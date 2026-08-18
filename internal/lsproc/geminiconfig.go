package lsproc

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const StandaloneTokenFile = "jetski-standalone-oauth-token"

// TokenPath is where the standalone OAuth token lives.
func TokenPath(geminiDir string) string {
	return filepath.Join(geminiDir, StandaloneTokenFile)
}

// HasStandaloneToken reports whether this machine has signed in to Antigravity.
// A remote server cannot complete the localhost OAuth callback, so the token
// usually has to be copied from a desktop machine.
func HasStandaloneToken(geminiDir string) bool {
	info, err := os.Stat(TokenPath(geminiDir))
	return err == nil && info.Size() > 0
}

var ErrConfigMissing = errors.New("antigravity settings file does not exist yet")

// EnableRemoteControl turns on remoteControlEnabled in Antigravity's settings,
// reporting whether anything changed. It returns ErrConfigMissing before
// Antigravity has ever run, which is not an error worth surfacing.
func EnableRemoteControl(geminiDir string) (changed bool, err error) {
	path := filepath.Join(geminiDir, "config", "config.json")

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, ErrConfigMissing
		}
		return false, err
	}

	var cfg map[string]any
	if err := json.Unmarshal(data, &cfg); err != nil {
		return false, fmt.Errorf("parse %s: %w", path, err)
	}

	settings, ok := cfg["userSettings"].(map[string]any)
	if !ok {
		settings = map[string]any{}
		cfg["userSettings"] = settings
	}

	if enabled, ok := settings["remoteControlEnabled"].(bool); ok && enabled {
		if _, ok := settings["cliRemoteControlHostname"]; ok {
			return false, nil
		}
	}

	settings["remoteControlEnabled"] = true

	if _, ok := settings["cliRemoteControlHostname"]; !ok {
		if hostname, ok := settings["remoteControlHostname"].(string); ok && hostname != "" {
			settings["cliRemoteControlHostname"] = hostname
		} else {
			h, _ := os.Hostname()
			settings["cliRemoteControlHostname"] = h
		}
	}

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, err
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		return false, err
	}
	return true, nil
}
