package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Config representa la configuración de pingbar
type Config struct {
	APIKey       string
	Lang         string
	DefaultCity  string
	Color        string
	DefaultLimit int
}

// ConfigDir devuelve el directorio de configuración según el SO
func ConfigDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("APPDATA"), "pingbar")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pingbar")
}

// ConfigFile devuelve la ruta del archivo de configuración
func ConfigFile() string {
	return filepath.Join(ConfigDir(), "config")
}

// CacheDir devuelve el directorio de caché
func CacheDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "pingbar", "cache")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "pingbar")
}

// Load carga la configuración desde el archivo
func Load() (*Config, error) {
	cfg := &Config{
		Lang:         "es",
		Color:        "auto",
		DefaultLimit: 10,
	}

	file, err := os.Open(ConfigFile())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "apikey":
			cfg.APIKey = value
		case "lang":
			cfg.Lang = value
		case "default-city":
			cfg.DefaultCity = value
		case "color":
			cfg.Color = value
		case "default-limit":
			var limit int
			fmt.Sscanf(value, "%d", &limit)
			if limit > 0 && limit <= 50 {
				cfg.DefaultLimit = limit
			}
		}
	}

	return cfg, scanner.Err()
}

// Set establece un valor de configuración
func Set(key, value string) error {
	validKeys := map[string]bool{
		"apikey":        true,
		"lang":          true,
		"default-city":  true,
		"color":         true,
		"default-limit": true,
	}

	if !validKeys[key] {
		return fmt.Errorf("clave de configuración no válida: %s", key)
	}

	switch key {
	case "lang":
		if value != "es" && value != "en" {
			return fmt.Errorf("idioma no válido: %s (usa 'es' o 'en')", value)
		}
	case "color":
		if value != "on" && value != "off" && value != "auto" {
			return fmt.Errorf("valor de color no válido: %s (usa 'on', 'off' o 'auto')", value)
		}
	case "default-limit":
		var limit int
		_, err := fmt.Sscanf(value, "%d", &limit)
		if err != nil || limit < 1 || limit > 50 {
			return fmt.Errorf("límite no válido: %s (debe ser entre 1 y 50)", value)
		}
	}

	if err := os.MkdirAll(ConfigDir(), 0755); err != nil {
		return err
	}

	existingConfig := make(map[string]string)
	file, err := os.Open(ConfigFile())
	if err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				existingConfig[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
		file.Close()
	}

	existingConfig[key] = value

	outFile, err := os.Create(ConfigFile())
	if err != nil {
		return err
	}
	defer outFile.Close()

	for k, v := range existingConfig {
		fmt.Fprintf(outFile, "%s=%s\n", k, v)
	}

	return nil
}

// Get obtiene un valor de configuración
func Get(key string) (string, error) {
	cfg, err := Load()
	if err != nil {
		return "", err
	}

	switch key {
	case "apikey":
		return cfg.APIKey, nil
	case "lang":
		return cfg.Lang, nil
	case "default-city":
		return cfg.DefaultCity, nil
	case "color":
		return cfg.Color, nil
	case "default-limit":
		return fmt.Sprintf("%d", cfg.DefaultLimit), nil
	default:
		return "", fmt.Errorf("clave de configuración no válida: %s", key)
	}
}

// List devuelve todas las configuraciones
func List() (map[string]string, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	result["apikey"] = maskAPIKey(cfg.APIKey)
	result["lang"] = cfg.Lang
	result["default-city"] = cfg.DefaultCity
	result["color"] = cfg.Color
	result["default-limit"] = fmt.Sprintf("%d", cfg.DefaultLimit)

	return result, nil
}

func maskAPIKey(key string) string {
	if key == "" {
		return "(no configurada)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// HasAPIKey verifica si hay una API key configurada
func HasAPIKey() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return cfg.APIKey != ""
}

// GetAPIKey devuelve la API key
func GetAPIKey() string {
	cfg, err := Load()
	if err != nil {
		return ""
	}
	return cfg.APIKey
}

