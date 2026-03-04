package cmd

import (
	"github.com/686f6c61/pingbar/internal/output"
	"github.com/spf13/cobra"
)

// aboutCmd muestra informacion sobre el programa
var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Mostrar autor, version y datos del programa",
	Long:  `Muestra informacion sobre pingbar, su autor y version.`,
	Run: func(cmd *cobra.Command, args []string) {
		output.PrintAbout(Version)
	},
}
