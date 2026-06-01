package cmd

import (
	"log"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	apiKey  string
	model   string
	verbose bool
	client  anthropic.Client
)

var rootCmd = &cobra.Command{
	Use:   "codeaid",
	Short: "A coding agent powered by Claude",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := godotenv.Load(); err != nil {
			log.Println("no .env file found, reading from environment directly")
		}
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			log.Fatal("ANTHROPIC_API_KEY is not set (use --api-key or set ANTHROPIC_API_KEY)")
		}
		client = anthropic.NewClient(option.WithAPIKey(apiKey))
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Anthropic API key (overrides ANTHROPIC_API_KEY)")
	rootCmd.PersistentFlags().StringVar(&model, "model", string(anthropic.ModelClaudeHaiku4_5), "Claude model to use")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Print token usage and response metadata")
}
