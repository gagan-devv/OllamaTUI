package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/gagan-devv/ollama-go/internal/ui"
	"github.com/gagan-devv/ollama-go/internal/config"
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

		// Load configuration
		configManager := config.NewConfigManager("")
		cfg, err := configManager.Load()
		if err != nil {
			fmt.Printf("Warning: Could not load config, using defaults: %v\n", err)
			cfg = config.GetDefault()
		}

		// Ensure test_output directory exists
		if err := os.MkdirAll(cfg.Paths.TestFolder, 0755); err != nil {
			fmt.Printf("Warning: Could not create test folder: %v\n", err)
		}

		// Use config values or command-line flags
		modelName := viper.GetString("model")
		if modelName == "" || modelName == "gemma3:270m" {
			modelName = cfg.Model.Default
		}
		
		systemPrompt := viper.GetString("system")
		if systemPrompt == "You are a helpful assistant." {
			systemPrompt = cfg.Model.SystemPrompt
		}

		m := ui.InitialModel(client, modelName, systemPrompt, cfg, configManager)
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