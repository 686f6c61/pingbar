package cmd

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/686f6c61/pingbar/internal/cache"
	"github.com/686f6c61/pingbar/internal/config"
	"github.com/686f6c61/pingbar/internal/i18n"
	"github.com/spf13/cobra"
)

// uninstallCmd desinstala pingbar del sistema
var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Desinstalar pingbar del sistema",
	Long: `Desinstala pingbar del sistema.

Este comando:
1. Solicita confirmación al usuario
2. Elimina el binario del sistema
3. Pregunta si desea eliminar la configuración
4. Pregunta si desea eliminar la caché`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.Load()
		msgs := i18n.Get(i18n.Lang(cfg.Lang))
		reader := bufio.NewReader(os.Stdin)

		// Confirmación principal
		fmt.Print(msgs.UninstallConfirm)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToUpper(answer))

		if answer != "Y" && answer != "S" {
			fmt.Println("Operación cancelada.")
			return
		}

		// Obtener ruta del ejecutable actual
		execPath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error al obtener ruta del ejecutable: %v\n", err)
			return
		}

		// Preguntar sobre configuración
		fmt.Print(msgs.DeleteConfig)
		answer, _ = reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToUpper(answer))
		deleteConfig := answer == "Y" || answer == "S"

		// Preguntar sobre caché
		fmt.Print(msgs.DeleteCache)
		answer, _ = reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToUpper(answer))
		deleteCache := answer == "Y" || answer == "S"

		// Eliminar configuración si se solicitó
		if deleteConfig {
			if err := os.RemoveAll(config.ConfigDir()); err != nil {
				fmt.Printf("Advertencia: no se pudo eliminar configuración: %v\n", err)
			}
		}

		// Eliminar caché si se solicitó
		if deleteCache {
			if err := cache.Clear(); err != nil {
				fmt.Printf("Advertencia: no se pudo eliminar caché: %v\n", err)
			}
			os.RemoveAll(config.CacheDir())
		}

		// Mostrar instrucciones para eliminar el binario
		// No podemos eliminar el binario mientras se está ejecutando
		fmt.Println()
		fmt.Println(msgs.UninstallDone)
		fmt.Println()
		fmt.Println("Para completar la desinstalación, ejecuta:")
		fmt.Println()

		if runtime.GOOS == "windows" {
			fmt.Printf("  del \"%s\"\n", execPath)
		} else {
			fmt.Printf("  sudo rm \"%s\"\n", execPath)
		}
	},
}

