package cmd

import (
	"fmt"
	"os"

	"github.com/686f6c61/pingbar/internal/api"
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

	// Buscar (incluye extracción de horarios de snippets)
	results, err := api.Search(cfg.APIKey, business, city, limit)
	if err != nil {
		if apiErr, ok := err.(*api.APIError); ok {
			output.PrintError(apiErr.Type, lang)
		} else {
			output.PrintError(err.Error(), lang)
		}
		os.Exit(1)
	}

	// Mostrar resultados
	formatter.PrintResults(results, business, city, showWeek)
}
