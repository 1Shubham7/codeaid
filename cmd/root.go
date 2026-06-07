package cmd

import (
	"os"

	"github.com/1shubham7/codeaid/logger"
	"github.com/1shubham7/codeaid/tools"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	apiKey  string
	model   string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "codeaid",
	Short: "A coding agent powered by Claude",
	Run: func(cmd *cobra.Command, args []string) {
		runTUI()
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.Init()
		godotenv.Load()
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}

		cfg := loadConfig()

		if !cmd.Root().PersistentFlags().Changed("model") && cfg.Model != "" {
			model = cfg.Model
		}

		// Seed defaults into config.json on first run so the user can customise them.
		changed := false
		if len(cfg.RestrictedCommands) == 0 {
			cfg.RestrictedCommands = tools.DefaultRestrictedCommands
			changed = true
		}
		if cfg.MaxFileSizeKB == 0 {
			cfg.MaxFileSizeKB = 100
			changed = true
		}
		if changed {
			saveConfig(cfg)
		}

		tools.SetRestrictedCommands(cfg.RestrictedCommands)
		tools.SetMaxFileSizeKB(cfg.MaxFileSizeKB)

		logger.L.Info("codeaid started", "model", model, "api_key_set", apiKey != "")
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
