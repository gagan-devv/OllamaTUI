package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/gagan-devv/ollama-go/internal/ui"
)

var rootCmd = &cobra.Command{
	Use:   "chat",
	Short: "Ollama CLI with TUI",
	Run: func(cmd *cobra.Command, args []string) {
		client, err := api.ClientFromEnvironment()
		if err != nil {
			fmt.Printf("Error connecting to Ollama: %v\n", err)
			os.Exit(1)
		}

		m := ui.InitialModel(client, viper.GetString("model"), viper.GetString("system"))
		p := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running program: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("model", "m", "qwen3.5:cloud", "Model name")
	rootCmd.PersistentFlags().StringP("system", "s", "You are a helpful assistant.", "System prompt")
	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("system", rootCmd.PersistentFlags().Lookup("system"))
}