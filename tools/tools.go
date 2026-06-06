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
	{OfTool: &anthropic.ToolParam{
		Name:        "execute_code",
		Description: anthropic.String("Executes a shell command and returns the exit code, stdout, and stderr. Times out after 30 seconds."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The shell command to execute.",
				},
			},
			Required: []string{"command"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "list_directory",
		Description: anthropic.String("Lists the contents of a directory, returning directories and files separately. Defaults to the current working directory if no path is given."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "Path to the directory to list. Defaults to '.' if omitted.",
				},
			},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "read_file",
		Description: anthropic.String("Reads and returns the contents of a file at the given path."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The path to the file to read.",
				},
			},
			Required: []string{"path"},
		},
	}},
	{OfTool: &anthropic.ToolParam{
		Name:        "write_file",
		Description: anthropic.String("Writes content to a file at the given path, creating any missing parent directories."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "The path to the file to write.",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The content to write to the file.",
				},
			},
			Required: []string{"path", "content"},
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
	case "execute_code":
		var input struct {
			Command string `json:"command"`
		}
		json.Unmarshal(rawInput, &input)
		return executeCode(input.Command)
	case "list_directory":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		return listDirectory(input.Path)
	case "read_file":
		var input struct {
			Path string `json:"path"`
		}
		json.Unmarshal(rawInput, &input)
		return readFile(input.Path)
	case "write_file":
		var input struct {
			Path    string `json:"path"`
			Content string `json:"content"`
		}
		json.Unmarshal(rawInput, &input)
		return writeFile(input.Path, input.Content)
	default:
		return fmt.Sprintf("unknown tool: %s", name)
	}
}