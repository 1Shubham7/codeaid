package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/spf13/cobra"
)

const historyFile = ".codeaid/history.json"

var codeCmd = &cobra.Command{
	Use:   "code",
	Short: "Start an interactive coding session with memory",
	Run:   runChat,
}

func init() {
	rootCmd.AddCommand(codeCmd)
}

func historyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot find home directory: %v", err)
	}
	return filepath.Join(home, historyFile)
}

func loadHistory() []anthropic.MessageParam {
	data, err := os.ReadFile(historyPath())
	if os.IsNotExist(err) {
		return []anthropic.MessageParam{}
	}
	if err != nil {
		log.Fatalf("failed to read history: %v", err)
	}
	var messages []anthropic.MessageParam
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Fatalf("failed to parse history: %v", err)
	}
	return messages
}

func saveHistory(messages []anthropic.MessageParam) {
	path := historyPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		log.Fatalf("failed to create history directory: %v", err)
	}
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("failed to serialize history: %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Fatalf("failed to save history: %v", err)
	}
}

func extractText(message *anthropic.Message) string {
	var sb strings.Builder
	for _, block := range message.Content {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			sb.WriteString(text.Text)
		}
	}
	return sb.String()
}

func runChat(_ *cobra.Command, _ []string) {
	messages := loadHistory()
	fmt.Printf("codeaid - %d messages loaded from history\n", len(messages))
	fmt.Println("type 'exit' to quit, 'clear' to wipe history")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("you: ")
		if !scanner.Scan() {
			break
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}
		if input == "exit" {
			break
		}
		if input == "clear" {
			messages = []anthropic.MessageParam{}
			saveHistory(messages)
			fmt.Println("history cleared")
			continue
		}

		messages = append(messages, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

		resp, err := client.Messages.New(context.Background(), anthropic.MessageNewParams{
			Model:     model,
			MaxTokens: 1024,
			Messages:  messages,
		})
		if err != nil {
			log.Printf("API error: %v", err)
			messages = messages[:len(messages)-1]
			continue
		}

		reply := extractText(resp)
		messages = append(messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(reply)))
		saveHistory(messages)

		fmt.Printf("\ncodeaid: %s\n\n", reply)
		if verbose {
			fmt.Printf("[model: %s | stop: %s | tokens in: %d, out: %d, total: %d]\n\n",
				resp.Model,
				resp.StopReason,
				resp.Usage.InputTokens,
				resp.Usage.OutputTokens,
				resp.Usage.InputTokens+resp.Usage.OutputTokens,
			)
		}
	}
}
