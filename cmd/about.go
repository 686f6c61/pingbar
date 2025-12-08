package cmd

import (
	"github.com/686f6c61/pingbar/internal/output"
	"github.com/spf13/cobra"
)

// aboutCmd muestra información sobre el programa
var aboutCmd = &cobra.Command{
	Use:   "about",
	Short: "Mostrar autor, versión y año",
	Long:  `Muestra información sobre pingbar, su autor y versión.`,
	Run: func(cmd *cobra.Command, args []string) {
		output.PrintAbout()
	},
}

