package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	cfg := Load()
	if cfg.APIAddr == "" {
		t.Error("expected non-empty APIAddr")
	}
	if cfg.MCPAddr == "" {
		t.Error("expected non-empty MCPAddr")
	}
	if cfg.AppEnv == "" {
		t.Error("expected non-empty AppEnv")
	}
}
