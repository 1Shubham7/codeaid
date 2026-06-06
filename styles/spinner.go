package styles

import (
	"time"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/spinner"
)

var SpinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("205")) // pink/magenta

var MiddleFinger = spinner.Spinner{
	Frames: []string{
        "🖕         ",
        " 🖕        ",
        "  🖕       ",
        "   🖕      ",
        "    🖕     ",
        "     🖕    ",
        "      🖕   ",
        "       🖕  ",
        "        🖕 ",
        "         🖕",
		"         🖕",
		"        🖕 ",
		"       🖕  ",
		"      🖕   ",
		"     🖕    ",
		"    🖕     ",
		"   🖕      ",
		"  🖕       ",
		" 🖕        ",
		"🖕         ",
    },
    FPS: time.Second / 10,
}