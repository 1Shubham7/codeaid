package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/1shubham7/codeaid/tools"
	"github.com/anthropics/anthropic-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

//go:embed system_prompt.md
var systemPromptText string

// ToolCall records a single tool invocation for display in the TUI.
type ToolCall struct {
	Display string // human-readable summary shown before Claude's reply
}

// ResponseMsg is returned to the TUI as a tea.Msg once the full agentic loop completes.
type ResponseMsg struct {
	Reply        string
	ToolCalls    []ToolCall
	InputTokens  int64
	OutputTokens int64
	ModelUsed    string
	StopReason   string
	Err          error
}

// CallAPI runs the agentic loop in a goroutine. It handles tool calls automatically:
// if Claude returns stop_reason "tool_use", it calls the tool and loops until Claude
// produces a final text response, then returns ResponseMsg to the TUI.
func CallAPI(c anthropic.Client, messages []anthropic.MessageParam, model string) tea.Cmd {
	return func() tea.Msg {
		msgs := make([]anthropic.MessageParam, len(messages))
		copy(msgs, messages)

		var totalIn, totalOut int64
		var usedTools []ToolCall

		for {
			resp, err := c.Messages.New(context.Background(), anthropic.MessageNewParams{
				Model:     model,
				MaxTokens: 1024,
				Messages:  msgs,
				Tools:     tools.Definitions,
				System:    []anthropic.TextBlockParam{{Text: systemPromptText}},
			})
			if err != nil {
				return ResponseMsg{Err: err}
			}

			totalIn += resp.Usage.InputTokens
			totalOut += resp.Usage.OutputTokens

			if resp.StopReason != anthropic.StopReasonToolUse {
				return ResponseMsg{
					Reply:        extractText(resp),
					ToolCalls:    usedTools,
					InputTokens:  totalIn,
					OutputTokens: totalOut,
					ModelUsed:    resp.Model,
					StopReason:   string(resp.StopReason),
				}
			}

			// Claude wants to call tools — build the full assistant turn first
			var assistantBlocks []anthropic.ContentBlockParamUnion
			var toolCalls []anthropic.ToolUseBlock

			for _, block := range resp.Content {
				switch v := block.AsAny().(type) {
				case anthropic.TextBlock:
					assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(v.Text))
				case anthropic.ToolUseBlock:
					assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(v.ID, v.Input, v.Name))
					toolCalls = append(toolCalls, v)
				}
			}
			msgs = append(msgs, anthropic.NewAssistantMessage(assistantBlocks...))

			// Call each tool, record a display summary, collect results
			var resultBlocks []anthropic.ContentBlockParamUnion
			for _, tc := range toolCalls {
				result := tools.Dispatch(tc.Name, tc.Input)
				resultBlocks = append(resultBlocks, anthropic.NewToolResultBlock(tc.ID, result, false))
				usedTools = append(usedTools, ToolCall{Display: toolSummary(tc.Name, tc.Input)})
			}
			msgs = append(msgs, anthropic.NewUserMessage(resultBlocks...))
		}
	}
}

// toolSummary returns a short human-readable line for each tool call.
func toolSummary(name string, rawInput json.RawMessage) string {
	switch name {
	case "read_file":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		return "successfully read the file contents for file: " + input.Path
	case "get_current_time":
		return "fetched current time"
	default:
		return "called tool: " + name
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
