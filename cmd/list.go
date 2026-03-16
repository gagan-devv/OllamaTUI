package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/ollama/ollama/api"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all downloaded Ollama models",
	Run: func(cmd *cobra.Command, args []string) {
		client, _ := api.ClientFromEnvironment()
		resp, err := client.List(context.Background())
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		fmt.Println("Available Models:")
		for _, m := range resp.Models {
			fmt.Printf(" - %s (%s)\n", m.Name, m.Details.Format)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}