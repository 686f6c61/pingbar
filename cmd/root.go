package cmd

import (
	"fmt"
	"os"

	"github.com/686f6c61/pingbar/internal/config"
	"github.com/686f6c61/pingbar/internal/output"
	"github.com/spf13/cobra"
)

var (
	// Flags globales
	jsonOutput bool
	showWeek   bool
	showTomorrow bool
	langFlag   string
	noColor    bool
	limitFlag  int

	// Versión
	Version = "0.0.1"
)

// rootCmd representa el comando base
var rootCmd = &cobra.Command{
	Use:   "pingbar <negocio> <ciudad>",
	Short: "Consulta horarios comerciales de negocios",
	Long: `pingbar es una herramienta de línea de comandos que consulta 
el horario comercial de cualquier negocio indexado en Google.

En lugar de devolver una IP como el comando ping, 
devuelve si el establecimiento está abierto o cerrado, junto con su horario.

Ejemplos:
  pingbar "el corte ingles" madrid
  pingbar "farmacia" madrid
  pingbar "mercadona" barcelona`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		// Si no hay argumentos, mostrar ayuda o mensaje de bienvenida
		if len(args) == 0 {
			// Verificar si hay API key configurada
			if !config.HasAPIKey() {
				cfg, _ := config.Load()
				output.PrintWelcome(cfg.Lang)
				return
			}
			cmd.Help()
			return
		}

		// Necesitamos al menos negocio y ciudad
		if len(args) < 2 {
			// Verificar si hay ciudad por defecto
			cfg, _ := config.Load()
			if cfg.DefaultCity == "" {
				fmt.Println("Uso: pingbar <negocio> <ciudad>")
				fmt.Println("O configura una ciudad por defecto: pingbar config set default-city <ciudad>")
				os.Exit(1)
			}
			// Usar ciudad por defecto
			runSearch(args[0], cfg.DefaultCity)
			return
		}

		runSearch(args[0], args[1])
	},
}

// Execute ejecuta el comando raíz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Flags globales
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Salida en formato JSON")
	rootCmd.PersistentFlags().BoolVar(&showWeek, "week", false, "Mostrar horario completo de la semana")
	rootCmd.PersistentFlags().BoolVar(&showTomorrow, "tomorrow", false, "Mostrar horario de mañana")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Idioma de salida (es|en)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Desactivar colores en la salida")
	rootCmd.PersistentFlags().IntVar(&limitFlag, "limit", 0, "Limitar número de resultados (máximo 50)")

	// Añadir subcomandos
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(cacheCmd)
	rootCmd.AddCommand(aboutCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(versionCmd)
}

// versionCmd muestra la versión
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostrar versión",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pingbar v%s\n", Version)
	},
}

