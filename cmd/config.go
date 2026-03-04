package cmd

import (
	"fmt"

	"github.com/686f6c61/pingbar/internal/config"
	"github.com/686f6c61/pingbar/internal/i18n"
	"github.com/spf13/cobra"
)

// configCmd es el comando principal de configuracion
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Gestionar configuracion",
	Long:  `Gestionar la configuracion de pingbar: API key, idioma, ciudad por defecto, etc.`,
}

// configSetCmd establece un valor de configuracion
var configSetCmd = &cobra.Command{
	Use:   "set <clave> <valor>",
	Short: "Establecer valor de configuracion",
	Long: `Establecer un valor de configuracion.

Claves disponibles:
  apikey        - API Key de Serper.dev (obligatorio)
  lang          - Idioma de salida (es/en)
  default-city  - Ciudad por defecto para busquedas
  color         - Colores en terminal (on/off/auto)
  default-limit - Numero de resultados por defecto (1-50)

Ejemplos:
  pingbar config set apikey XXXXXXXXXXXXXXXXXXXX
  pingbar config set lang es
  pingbar config set default-city sevilla`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		value := args[1]

		err := config.Set(key, value)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		cfg, _ := config.Load()
		msgs := i18n.Get(i18n.Lang(cfg.Lang))
		displayValue := value
		if key == "apikey" {
			displayValue = config.MaskAPIKey(value)
		}
		fmt.Printf(msgs.ConfigSet+"\n", key, displayValue)
	},
}

// configGetCmd obtiene un valor de configuracion
var configGetCmd = &cobra.Command{
	Use:   "get <clave>",
	Short: "Obtener valor de configuracion",
	Long: `Obtener un valor de configuracion especifico.

Claves disponibles:
  apikey, lang, default-city, color, default-limit

Ejemplo:
  pingbar config get lang`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		value, err := config.Get(key)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		cfg, _ := config.Load()
		msgs := i18n.Get(i18n.Lang(cfg.Lang))

		displayValue := value
		if key == "apikey" {
			displayValue = config.MaskAPIKey(value)
		}

		fmt.Printf(msgs.ConfigGet+"\n", key, displayValue)
	},
}

// configListCmd lista toda la configuracion
var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Mostrar toda la configuracion",
	Run: func(cmd *cobra.Command, args []string) {
		configMap, err := config.List()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Println("Configuracion actual:")
		fmt.Println()

		keys := []string{"apikey", "lang", "default-city", "color", "default-limit"}
		for _, key := range keys {
			value := configMap[key]
			if value == "" {
				value = "(no configurado)"
			}
			fmt.Printf("  %-14s = %s\n", key, value)
		}

		fmt.Println()
		fmt.Printf("Archivo de configuracion: %s\n", config.ConfigFile())
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}
