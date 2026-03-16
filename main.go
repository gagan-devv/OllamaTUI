/*
Copyright © 2026 Gagan Ahlawat - gagan.devvv@gmail.com
*/
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	// "github.com/gagan-devv/ollama-go/cmd"
	"github.com/ollama/ollama/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	history []api.Message
	model   string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ollama-chat",
		Short: "A stateful Ollama CLI client",
		Run:   runChat,
	}

	// Viper Configuration
	viper.SetDefault("model", "qwen3.5:cloud")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.ReadInConfig()

	model = viper.GetString("model")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startOllama() {
	cmd := exec.Command("ollama", "serve")
	if err := cmd.Start(); err != nil {
		log.Printf("Note: Could not start Ollama: %v", err)
	}
	time.Sleep(2 * time.Second)
}

func runChat(cmd *cobra.Command, args []string) {
	startOllama()
	client, _ := api.ClientFromEnvironment()
	ctx := context.Background()
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Using model: %s. Type 'bye' to exit.\n", model)

	for {
		fmt.Print("\033[32m👤 You: \033[0m")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()

		if strings.ToLower(input) == "bye" {
			break
		}

		// Add User input to History (Context)
		userMsg := api.Message{Role: "user", Content: input}
		history = append(history, userMsg)

		req := &api.ChatRequest{
			Model:    model,
			Messages: history,
		}

		fmt.Print("\033[33m🤖 AI: \033[0m")
		var fullResponse strings.Builder

		// Streaming response logic
		fn := func(resp api.ChatResponse) error {
			fmt.Print(resp.Message.Content)
			fullResponse.WriteString(resp.Message.Content)
			return nil
		}

		err := client.Chat(ctx, req, fn)
		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			continue
		}
		fmt.Print("\n\n")

		history = append(history, api.Message{
			Role:    "assistant",
			Content: fullResponse.String(),
		})
	}
}
