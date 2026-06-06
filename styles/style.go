package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// toolStyle is the lipgloss style applied to tool-call status lines.
// lipgloss.NewStyle() starts a blank style — you chain methods to build it up.
// .Foreground() sets the text colour; lipgloss.Color() accepts:
//   - ANSI 256-colour codes ("34" = blue, "2" = green, "11" = yellow)
//   - True-colour hex strings ("#04B575")
//
// .Italic(true) makes the text italic.
// .Render(s) applies the style and returns the styled string.
var ToolStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("2")). // ANSI green
	Italic(true)

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("205")) // pink/magenta