package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all downloaded Ollama models",
	Long:  `Retrieve a list of all models currently available in your local Ollama instance.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 1. Connect to Ollama
		client, err := api.ClientFromEnvironment()
		if err != nil {
			fmt.Printf("Error connecting to Ollama: %v\n", err)
			return
		}

		// 2. Fetch models
		ctx := context.Background()
		resp, err := client.List(ctx)
		if err != nil {
			fmt.Printf("Error fetching models: %v\n", err)
			return
		}

		if len(resp.Models) == 0 {
			fmt.Println("No models found. Use 'ollama pull' to download some.")
			return
		}

		// 3. Pretty print using tabwriter for a clean terminal table
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tID\tSIZE\tMODIFIED")
		
		for _, m := range resp.Models {
			// Convert bytes to GB for readability
			sizeGB := float64(m.Size) / (1024 * 1024 * 1024)
			fmt.Fprintf(w, "%s\t%s\t%.2f GB\t%s\n", 
				m.Name, 
				m.Digest[:12], // Short digest
				sizeGB, 
				m.ModifiedAt.Format("2 Jan 2006"),
			)
		}
		w.Flush()
	},
}

func init() {
	// This line is what allows './chat list' to work
	rootCmd.AddCommand(listCmd)
}