package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	model        string
	systemPrompt string
	history      []api.Message
)

var rootCmd = &cobra.Command{
	Use:   "chat",
	Short: "Stateful AI Chat",
	Run:   runChat,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Flags
	rootCmd.PersistentFlags().StringVarP(&model, "model", "m", "qwen3.5", "Ollama model to use")
	rootCmd.PersistentFlags().StringVarP(&systemPrompt, "system", "s", "You are a helpful assistant.", "System instructions for the AI.")

	viper.BindPFlag("model", rootCmd.PersistentFlags().Lookup("model"))
	viper.BindPFlag("system", rootCmd.PersistentFlags().Lookup("system"))
}

func runChat(cmd *cobra.Command, args []string) {
	client, _ := api.ClientFromEnvironment()
	scanner := bufio.NewScanner(os.Stdin)

	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Bold(true)

	fmt.Println("Chat Started. Type 'bye' to quit.")

	for {
		fmt.Print(userStyle.Render("👤 You: "))
		if !scanner.Scan() {
			break
		}

		history = append(history, api.Message{
			Role: "system",
			Content: viper.GetString("system"),
		})

		input := scanner.Text()

		if strings.ToLower(input) == "bye" {
			break
		}

		history = append(history, api.Message{Role: "user", Content: input})

		if len(history) > 10 {
			history = history[len(history)-10:]
		}

		fmt.Print("\033[33m🤖 AI: \033[0m")
		var fullResponse strings.Builder

		req := &api.ChatRequest{Model: model, Messages: history}
		fn := func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)
			fullResponse.WriteString(resp.Message.Content)
			return nil
		}

		client.Chat(context.Background(), req, fn)
		fmt.Println("\n---")

		// Final Pretty Print of the full response using Glamour
		rendered, _ := glamour.Render(fullResponse.String(), "dark")
		fmt.Print(rendered)

		history = append(history, api.Message{Role: "assistant", Content: fullResponse.String()})
	}
}
