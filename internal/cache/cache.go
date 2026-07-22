package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/686f6c61/pingbar/internal/config"
)

// CacheEntry representa una entrada de caché
type CacheEntry struct {
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	TTLHours  int             `json:"ttl_hours"`
}

// DefaultTTL es el tiempo de vida por defecto de la caché (24 horas)
const DefaultTTL = 24

func generateKey(business, city string, limit int) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%s|%d", business, city, limit)
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func getCacheFile(key string) string {
	return filepath.Join(config.CacheDir(), key+".json")
}

// Get obtiene datos de la caché si existen y no han expirado
func Get(business, city string, limit int) (json.RawMessage, bool) {
	key := generateKey(business, city, limit)
	cacheFile := getCacheFile(key)

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	expirationTime := entry.Timestamp.Add(time.Duration(entry.TTLHours) * time.Hour)
	if time.Now().After(expirationTime) {
		os.Remove(cacheFile)
		return nil, false
	}

	return entry.Data, true
}

// Set guarda datos en la caché
func Set(business, city string, limit int, data json.RawMessage, ttlHours int) error {
	if ttlHours <= 0 {
		ttlHours = DefaultTTL
	}

	if err := os.MkdirAll(config.CacheDir(), 0700); err != nil {
		return err
	}

	key := generateKey(business, city, limit)
	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTLHours:  ttlHours,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(getCacheFile(key), jsonData, 0600)
}

// Clear limpia toda la caché
func Clear() error {
	cacheDir := config.CacheDir()

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil
	}

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			os.Remove(filepath.Join(cacheDir, entry.Name()))
		}
	}

	return nil
}

// Size devuelve el número de entradas en la caché
func Size() int {
	cacheDir := config.CacheDir()

	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return 0
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			count++
		}
	}

	return count
}
