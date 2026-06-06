package tools

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
)

// Definitions is passed to every API call so Claude knows what tools exist.
var Definitions = []anthropic.ToolUnionParam{
	{OfTool: &anthropic.ToolParam{
		Name:        "get_current_time",
		Description: anthropic.String("Returns the current date and time. Use this whenever the user asks about the current time, date, or day."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"timezone": map[string]any{
					"type":        "string",
					"description": "Timezone name e.g. 'UTC', 'America/New_York'. Defaults to local time if omitted.",
				},
			},
		},
	}},
}

// Dispatch routes a tool_use call from Claude to the right Go function.
func Dispatch(name string, rawInput json.RawMessage) string {
	switch name {
	case "get_current_time":
		var input struct {
			Timezone string `json:"timezone"`
		}
		json.Unmarshal(rawInput, &input)
		return getCurrentTime(input.Timezone)
	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}