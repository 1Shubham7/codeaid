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

var ExecOkStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("12")) // bright blue

var ExecErrStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")) // bright red

var CoolArrow = "❯ "

var CoolArrowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

var InputBoxStyle = lipgloss.NewStyle().
	Border(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("205")).
	Padding(0, 1)

var CodaidLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("205")).
	Bold(true)

var YouLabelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("12")).
	Background(lipgloss.Color("225")).
	Bold(true)

var YouBoxStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("225")).
	Foreground(lipgloss.Color("22")).
	Bold(true).
	Padding(0, 1)


