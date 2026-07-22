package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/686f6c61/pingbar/internal/api"
	"github.com/686f6c61/pingbar/internal/cache"
	"github.com/686f6c61/pingbar/internal/config"
	"github.com/686f6c61/pingbar/internal/output"
)

// runSearch ejecuta la búsqueda principal
func runSearch(business, city string) {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error al cargar configuración: %v\n", err)
		os.Exit(1)
	}

	// Verificar API key
	if cfg.APIKey == "" {
		output.PrintWelcome(cfg.Lang)
		os.Exit(1)
	}

	// Determinar idioma
	lang := cfg.Lang
	if langFlag != "" {
		lang = langFlag
	}

	// Determinar modo de color
	colorMode := cfg.Color
	if noColor {
		colorMode = "off"
	}

	// Determinar límite
	limit := cfg.DefaultLimit
	if limitFlag > 0 {
		if limitFlag > 50 {
			limit = 50
		} else {
			limit = limitFlag
		}
	}

	// Crear formateador de salida
	formatter := output.NewFormatter(lang, colorMode, jsonOutput)

	results, ok := loadFromCache(business, city, limit)
	if !ok {
		// Buscar en la API (incluye extracción de horarios de snippets)
		results, err = api.Search(cfg.APIKey, business, city, limit)
		if err != nil {
			var apiErr *api.APIError
			if errors.As(err, &apiErr) {
				output.PrintError(apiErr.Type, lang)
			} else {
				output.PrintError(err.Error(), lang)
			}
			os.Exit(1)
		}
		saveToCache(business, city, limit, results)
	}

	// Mostrar resultados
	formatter.PrintResults(results, business, city)
}

// loadFromCache recupera resultados cacheados y recalcula el estado
// abierto/cerrado, que depende de la hora actual
func loadFromCache(business, city string, limit int) ([]api.BusinessInfo, bool) {
	data, ok := cache.Get(business, city, limit)
	if !ok {
		return nil, false
	}

	var results []api.BusinessInfo
	if err := json.Unmarshal(data, &results); err != nil {
		return nil, false
	}

	for i := range results {
		results[i].RefreshOpenState()
	}

	return results, true
}

// saveToCache guarda los resultados; un fallo de caché no interrumpe la búsqueda
func saveToCache(business, city string, limit int, results []api.BusinessInfo) {
	data, err := json.Marshal(results)
	if err != nil {
		return
	}
	if err := cache.Set(business, city, limit, data, cache.DefaultTTL); err != nil {
		fmt.Fprintf(os.Stderr, "Advertencia: no se pudo guardar la caché: %v\n", err)
	}
}
