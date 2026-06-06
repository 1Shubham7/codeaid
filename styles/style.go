package styles

import (
	"github.com/charmbracelet/lipgloss"
)

var ToolStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("2")).
	Italic(true)

var BlockToolCallStyle = lipgloss.NewStyle().
	Border(lipgloss.BlockBorder()).
	Padding(1, 2)
