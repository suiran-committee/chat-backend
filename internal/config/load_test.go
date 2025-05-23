package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("FRONTEND_ORIGIN", "")
	cfg := Load()

	if cfg.Port != "8443" {
		t.Errorf("default port = %s, want 8443", cfg.Port)
	}
	if cfg.RedisAddr != "localhost:6379" {
		t.Errorf("default redis = %s", cfg.RedisAddr)
	}
}
