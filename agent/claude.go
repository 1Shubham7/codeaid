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
	Display string
	Output  string // non-empty for execute_code — the raw stdout/stderr/exit output
	IsError bool   // true when exit code != 0
}

// IterationMsg is sent to the TUI after each tool-use round-trip in the agentic loop.
type IterationMsg struct {
	InputTokens  int64
	OutputTokens int64
	StopReason   string
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

// CallAPI runs the agentic loop in a goroutine. After each tool-use round-trip it sends
// an IterationMsg on iterCh so the TUI can show per-call token counts live. The channel
// is closed when the loop exits, signalling the TUI to stop listening.
func CallAPI(c anthropic.Client, messages []anthropic.MessageParam, model string, iterCh chan<- IterationMsg) tea.Cmd {
	return func() tea.Msg {
		defer close(iterCh)

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
				usedTools = append(usedTools, toolSummary(tc.Name, tc.Input, result))
			}
			msgs = append(msgs, anthropic.NewUserMessage(resultBlocks...))

			// Notify the TUI about this iteration's token usage.
			iterCh <- IterationMsg{
				InputTokens:  resp.Usage.InputTokens,
				OutputTokens: resp.Usage.OutputTokens,
				StopReason:   string(resp.StopReason),
			}
		}
	}
}

// toolSummary builds the ToolCall shown in the TUI for each tool invocation.
// result is the raw string returned by tools.Dispatch — used by execute_code
// to carry output and determine success/failure.
func toolSummary(name string, rawInput json.RawMessage, result string) ToolCall {
	switch name {
	case "execute_code":
		var input struct {
			Command string `json:"command"`
		}
		json.Unmarshal(rawInput, &input)
		return ToolCall{
			Display: "ran: " + input.Command,
			Output:  result,
			IsError: !strings.HasPrefix(result, "Exit Code: 0\n"),
		}
	case "list_directory":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		if input.Path == "" {
			return ToolCall{Display: "listed directory: ."}
		}
		return ToolCall{Display: "listed directory: " + input.Path}
	case "read_file":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		return ToolCall{Display: "read file: " + input.Path}
	case "write_file":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		return ToolCall{Display: "wrote file: " + input.Path}
	case "get_current_time":
		return ToolCall{Display: "fetched current time"}
	default:
		return ToolCall{Display: "called tool: " + name}
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
