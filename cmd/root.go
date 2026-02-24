package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fornax",
	Short: "Media processing CLI",
}

func Execute() error {
	return rootCmd.Execute()
}
