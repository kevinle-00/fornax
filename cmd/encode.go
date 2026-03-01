package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/kevinle-00/fornax/internal/encode"
	"github.com/kevinle-00/fornax/internal/validate"
	"github.com/spf13/cobra"
)

var encodeCmd = &cobra.Command{
	Use:   "encode <input> <output>",
	Short: "Encode media file format",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		inputPath := args[0]
		outputPath := args[1]
		if err := validate.IsValidInputPath(inputPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if err := validate.IsValidOutputPath(outputPath); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		// TODO: Make encode manually cancellable by user
		encoder := encode.New()
		err := encoder.Encode(context.Background(), inputPath, outputPath)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(encodeCmd)
}
