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
	langFlag   string
	noColor    bool
	limitFlag  int

	// Version
	Version = "0.1.0"
)

// rootCmd representa el comando base
var rootCmd = &cobra.Command{
	Use:   "pingbar <negocio> <ciudad>",
	Short: "Consulta horarios comerciales de negocios",
	Long: `pingbar es una herramienta de linea de comandos que consulta
el horario comercial de cualquier negocio indexado en Google.

En lugar de devolver una IP como el comando ping,
devuelve si el establecimiento esta abierto o cerrado, junto con su horario.

Ejemplos:
  pingbar "el corte ingles" madrid
  pingbar "farmacia" madrid
  pingbar "mercadona" barcelona`,
	Args: cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			if !config.HasAPIKey() {
				cfg, err := config.Load()
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error al cargar configuracion: %v\n", err)
					os.Exit(1)
				}
				output.PrintWelcome(cfg.Lang)
				return
			}
			cmd.Help()
			return
		}

		if len(args) < 2 {
			cfg, err := config.Load()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error al cargar configuracion: %v\n", err)
				os.Exit(1)
			}
			if cfg.DefaultCity == "" {
				fmt.Println("Uso: pingbar <negocio> <ciudad>")
				fmt.Println("O configura una ciudad por defecto: pingbar config set default-city <ciudad>")
				os.Exit(1)
			}
			runSearch(args[0], cfg.DefaultCity)
			return
		}

		runSearch(args[0], args[1])
	},
}

// Execute ejecuta el comando raiz
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Salida en formato JSON")
	rootCmd.PersistentFlags().StringVar(&langFlag, "lang", "", "Idioma de salida (es|en)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Desactivar colores en la salida")
	rootCmd.PersistentFlags().IntVar(&limitFlag, "limit", 0, "Limitar numero de resultados (maximo 50)")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(cacheCmd)
	rootCmd.AddCommand(aboutCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(versionCmd)
}

// versionCmd muestra la version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Mostrar version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("pingbar v%s\n", Version)
	},
}
