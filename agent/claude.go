package agent

import (
	"context"
	_ "embed"
	"encoding/json"
	"sort"
	"strings"

	"github.com/1shubham7/codeaid/tools"
	"github.com/anthropics/anthropic-sdk-go"
	tea "github.com/charmbracelet/bubbletea"
)

// blockAccum holds the content accumulated for a single content block during streaming.
type blockAccum struct {
	blockType string // "text" or "tool_use"
	text      strings.Builder
	toolID    string
	toolName  string
	toolInput strings.Builder
}

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

// CallAPI runs the agentic loop in a goroutine. Text chunks are streamed to streamCh
// as they arrive so the TUI can display them live. After each tool-use round-trip an
// IterationMsg is sent on iterCh. Both channels are closed when the loop exits.
func CallAPI(c anthropic.Client, messages []anthropic.MessageParam, model string, iterCh chan<- IterationMsg, streamCh chan<- string) tea.Cmd {
	return func() tea.Msg {
		defer close(iterCh)
		defer close(streamCh)

		msgs := make([]anthropic.MessageParam, len(messages))
		copy(msgs, messages)

		var totalIn, totalOut int64
		var usedTools []ToolCall

		for {
			stream := c.Messages.NewStreaming(context.Background(), anthropic.MessageNewParams{
				Model:     model,
				MaxTokens: 1024,
				Messages:  msgs,
				Tools:     tools.Definitions,
				System:    []anthropic.TextBlockParam{{Text: systemPromptText}},
			})

			// Accumulate content blocks keyed by their stream index.
			blocks := make(map[int64]*blockAccum)
			var iterIn, iterOut int64
			var stopReason anthropic.StopReason
			var modelUsed string

			for stream.Next() {
				event := stream.Current()
				switch e := event.AsAny().(type) {
				case anthropic.MessageStartEvent:
					iterIn = e.Message.Usage.InputTokens
					modelUsed = string(e.Message.Model)
				case anthropic.ContentBlockStartEvent:
					switch b := e.ContentBlock.AsAny().(type) {
					case anthropic.TextBlock:
						blocks[e.Index] = &blockAccum{blockType: "text"}
					case anthropic.ToolUseBlock:
						blocks[e.Index] = &blockAccum{blockType: "tool_use", toolID: b.ID, toolName: b.Name}
					}
				case anthropic.ContentBlockDeltaEvent:
					blk := blocks[e.Index]
					if blk == nil {
						continue
					}
					switch d := e.Delta.AsAny().(type) {
					case anthropic.TextDelta:
						blk.text.WriteString(d.Text)
						streamCh <- d.Text
					case anthropic.InputJSONDelta:
						blk.toolInput.WriteString(d.PartialJSON)
					}
				case anthropic.MessageDeltaEvent:
					stopReason = e.Delta.StopReason
					iterOut = e.Usage.OutputTokens
				}
			}

			if err := stream.Err(); err != nil {
				return ResponseMsg{Err: err}
			}

			totalIn += iterIn
			totalOut += iterOut

			// Rebuild the assistant turn in block-index order.
			indices := make([]int64, 0, len(blocks))
			for idx := range blocks {
				indices = append(indices, idx)
			}
			sort.Slice(indices, func(i, j int) bool { return indices[i] < indices[j] })

			var assistantBlocks []anthropic.ContentBlockParamUnion
			var toolCalls []anthropic.ToolUseBlock
			var replyText strings.Builder

			for _, idx := range indices {
				blk := blocks[idx]
				switch blk.blockType {
				case "text":
					if blk.text.Len() > 0 {
						assistantBlocks = append(assistantBlocks, anthropic.NewTextBlock(blk.text.String()))
						replyText.WriteString(blk.text.String())
					}
				case "tool_use":
					raw := json.RawMessage(blk.toolInput.String())
					assistantBlocks = append(assistantBlocks, anthropic.NewToolUseBlock(blk.toolID, raw, blk.toolName))
					toolCalls = append(toolCalls, anthropic.ToolUseBlock{ID: blk.toolID, Name: blk.toolName, Input: raw})
				}
			}

			if stopReason != anthropic.StopReasonToolUse {
				return ResponseMsg{
					Reply:        replyText.String(),
					ToolCalls:    usedTools,
					InputTokens:  totalIn,
					OutputTokens: totalOut,
					ModelUsed:    modelUsed,
					StopReason:   string(stopReason),
				}
			}

			msgs = append(msgs, anthropic.NewAssistantMessage(assistantBlocks...))

			// Dispatch tools and collect results.
			var resultBlocks []anthropic.ContentBlockParamUnion
			for _, tc := range toolCalls {
				result := tools.Dispatch(tc.Name, tc.Input)
				resultBlocks = append(resultBlocks, anthropic.NewToolResultBlock(tc.ID, result, false))
				usedTools = append(usedTools, toolSummary(tc.Name, tc.Input, result))
			}
			msgs = append(msgs, anthropic.NewUserMessage(resultBlocks...))

			iterCh <- IterationMsg{
				InputTokens:  iterIn,
				OutputTokens: iterOut,
				StopReason:   string(stopReason),
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

