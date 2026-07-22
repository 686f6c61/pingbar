package config

import (
	"os"
	"testing"
)

func TestSetAndLoad(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	if err := Set("apikey", "test-key-1234567890"); err != nil {
		t.Fatalf("Set(apikey) fallo: %v", err)
	}
	if err := Set("lang", "en"); err != nil {
		t.Fatalf("Set(lang) fallo: %v", err)
	}
	if err := Set("default-limit", "5"); err != nil {
		t.Fatalf("Set(default-limit) fallo: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() fallo: %v", err)
	}

	if cfg.APIKey != "test-key-1234567890" {
		t.Errorf("APIKey = %q, esperado %q", cfg.APIKey, "test-key-1234567890")
	}
	if cfg.Lang != "en" {
		t.Errorf("Lang = %q, esperado %q", cfg.Lang, "en")
	}
	if cfg.DefaultLimit != 5 {
		t.Errorf("DefaultLimit = %d, esperado 5", cfg.DefaultLimit)
	}

	info, err := os.Stat(ConfigFile())
	if err != nil {
		t.Fatalf("no se pudo leer el archivo de configuracion: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("permisos del archivo = %o, esperado 0600", perm)
	}
}

func TestSetRejectsInvalidValues(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	invalid := [][2]string{
		{"lang", "fr"},
		{"color", "rainbow"},
		{"default-limit", "0"},
		{"default-limit", "51"},
		{"default-limit", "abc"},
		{"clave-inexistente", "valor"},
	}

	for _, kv := range invalid {
		if err := Set(kv[0], kv[1]); err == nil {
			t.Errorf("Set(%q, %q) deberia fallar", kv[0], kv[1])
		}
	}
}

func TestLoadDefaults(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() fallo: %v", err)
	}
	if cfg.Lang != "es" || cfg.Color != "auto" || cfg.DefaultLimit != 10 {
		t.Errorf("valores por defecto incorrectos: %+v", cfg)
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "(no configurada)"},
		{"corta", "****"},
		{"abcd1234efgh5678", "abcd...5678"},
	}
	for _, tt := range tests {
		if got := MaskAPIKey(tt.in); got != tt.want {
			t.Errorf("MaskAPIKey(%q) = %q, esperado %q", tt.in, got, tt.want)
		}
	}
}
