package cmd

import (
	"fmt"

	"github.com/686f6c61/pingbar/internal/cache"
	"github.com/686f6c61/pingbar/internal/config"
	"github.com/686f6c61/pingbar/internal/i18n"
	"github.com/spf13/cobra"
)

// cacheCmd es el comando principal de caché
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Gestionar caché local",
	Long:  `Gestionar la caché local de pingbar.`,
}

// cacheClearCmd limpia la caché
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Limpiar caché local",
	Long: `Elimina todos los datos almacenados en la caché local.

La caché almacena los horarios consultados para reducir llamadas a la API.
Por defecto, los datos se guardan durante 24 horas.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := cache.Clear()
		if err != nil {
			fmt.Printf("Error al limpiar caché: %v\n", err)
			return
		}

		cfg, _ := config.Load()
		msgs := i18n.Get(i18n.Lang(cfg.Lang))
		fmt.Println(msgs.CacheCleared)
	},
}

// cacheInfoCmd muestra información de la caché
var cacheInfoCmd = &cobra.Command{
	Use:   "info",
	Short: "Mostrar información de la caché",
	Run: func(cmd *cobra.Command, args []string) {
		size := cache.Size()
		cacheDir := config.CacheDir()

		fmt.Println("Información de caché:")
		fmt.Println()
		fmt.Printf("  Directorio: %s\n", cacheDir)
		fmt.Printf("  Entradas:   %d\n", size)
		fmt.Printf("  TTL:        %d horas\n", cache.DefaultTTL)
	},
}

func init() {
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheInfoCmd)
}

