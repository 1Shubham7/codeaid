package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/1shubham7/codeaid/agent"
	"github.com/1shubham7/codeaid/styles"
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	historyFile = ".codeaid/history.json"
	configFile  = ".codeaid/config.json"
)

// --- config ---

type appConfig struct {
	Model string `json:"model"`
}

func codeaidDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot find home directory: %v", err)
	}
	return filepath.Join(home, ".codeaid")
}

func loadConfig() appConfig {
	data, err := os.ReadFile(filepath.Join(codeaidDir(), "config.json"))
	if os.IsNotExist(err) {
		return appConfig{}
	}
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	var cfg appConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}

func saveConfig(cfg appConfig) {
	dir := codeaidDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create config directory: %v", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		log.Fatalf("failed to serialize config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0644); err != nil {
		log.Fatalf("failed to save config: %v", err)
	}
}

// --- history ---

func loadHistory() []anthropic.MessageParam {
	data, err := os.ReadFile(filepath.Join(codeaidDir(), "history.json"))
	if os.IsNotExist(err) {
		return []anthropic.MessageParam{}
	}
	if err != nil {
		log.Fatalf("failed to read history: %v", err)
	}
	var messages []anthropic.MessageParam
	if err := json.Unmarshal(data, &messages); err != nil {
		log.Fatalf("failed to parse history: %v", err)
	}
	return messages
}

func saveHistory(messages []anthropic.MessageParam) {
	dir := codeaidDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("failed to create history directory: %v", err)
	}
	data, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		log.Fatalf("failed to serialize history: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "history.json"), data, 0644); err != nil {
		log.Fatalf("failed to save history: %v", err)
	}
}

// --- TUI ---

type tuiState int

const (
	stateAPIKey tuiState = iota
	stateMenu
	stateModelSelect
	stateCode
	stateWaiting
)

type menuItem struct {
	label string
	desc  string
}

var menuItems = []menuItem{
	{"Code", "Start a coding session"},
	{"Load History", "Browse past conversations"},
	{"Model", "Change the active model"},
	{"Exit", "Quit codeaid"},
}

var availableModels = []struct {
	id   string
	desc string
}{
	{string(anthropic.ModelClaudeHaiku4_5), "Fast & affordable"},
	{string(anthropic.ModelClaudeSonnet4_6), "Balanced performance"},
	{string(anthropic.ModelClaudeOpus4_8), "Most capable"},
}

type entry struct {
	role string // "you", "codeaid", "meta"
	text string
}

type iterDoneMsg struct{}
type streamChunkMsg string
type streamDoneMsg struct{}

func waitForIteration(ch <-chan agent.IterationMsg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return iterDoneMsg{}
		}
		return msg
	}
}

func waitForChunk(ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return streamDoneMsg{}
		}
		return streamChunkMsg(chunk)
	}
}

type tuiModel struct {
	state        tuiState
	cursor       int
	modelCursor  int
	input        textinput.Model
	spin         spinner.Model
	logoStr      string
	iterCh       chan agent.IterationMsg
	streamCh     chan string
	streamBuf    string
	entries      []entry
	messages     []anthropic.MessageParam
	historyCount int
	client       anthropic.Client
	width        int
	errMsg       string
}

func newTUI() tuiModel {
	ti := textinput.New()
	ti.CharLimit = 0

	s := spinner.New()
	s.Spinner = styles.MiddleFinger
	s.Style = styles.SpinnerStyle

	m := tuiModel{input: ti, spin: s}

	if apiKey == "" {
		m.state = stateAPIKey
		m.input.Placeholder = "Paste your Anthropic API key..."
		m.input.EchoMode = textinput.EchoPassword
		m.input.Focus()
	} else {
		m.state = stateMenu
		m.client = anthropic.NewClient(option.WithAPIKey(apiKey))
	}

	return m
}

func (m tuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {

		case stateAPIKey:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEnter:
				key := strings.TrimSpace(m.input.Value())
				if key == "" {
					return m, nil
				}
				apiKey = key
				m.client = anthropic.NewClient(option.WithAPIKey(apiKey))
				m.input.SetValue("")
				m.input.Blur()
				m.state = stateMenu
				return m, nil
			}

		case stateMenu:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyUp:
				if m.cursor > 0 {
					m.cursor--
				}
			case tea.KeyDown:
				if m.cursor < len(menuItems)-1 {
					m.cursor++
				}
			case tea.KeyEnter:
				return m.selectMenuItem()
			}
			switch msg.String() {
			case "1":
				m.cursor = 0
				return m.selectMenuItem()
			case "2":
				m.cursor = 1
				return m.selectMenuItem()
			case "3":
				m.cursor = 2
				return m.selectMenuItem()
			case "4":
				m.cursor = 3
				return m.selectMenuItem()
			}

		case stateModelSelect:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				m.state = stateMenu
				return m, nil
			case tea.KeyUp:
				if m.modelCursor > 0 {
					m.modelCursor--
				}
			case tea.KeyDown:
				if m.modelCursor < len(availableModels)-1 {
					m.modelCursor++
				}
			case tea.KeyEnter:
				selected := availableModels[m.modelCursor].id
				model = selected
				saveConfig(appConfig{Model: selected})
				m.errMsg = fmt.Sprintf("model set to %s", selected)
				m.state = stateMenu
				return m, nil
			}

		case stateCode, stateWaiting:
			switch msg.Type {
			case tea.KeyCtrlC:
				return m, tea.Quit
			case tea.KeyEsc:
				m.input.Blur()
				m.state = stateMenu
				return m, nil
			case tea.KeyEnter:
				if m.state == stateWaiting {
					return m, nil
				}
				input := strings.TrimSpace(m.input.Value())
				if input == "" {
					return m, nil
				}
				m.input.SetValue("")
				if input == "exit" {
					m.input.Blur()
					m.state = stateMenu
					return m, nil
				}
				if input == "clear" {
					m.entries = []entry{{role: "tool", text: "screen cleared"}}
					return m, nil
				}
				if input == "clear history" {
					m.messages = []anthropic.MessageParam{}
					m.historyCount = 0
					saveHistory(m.messages)
					m.entries = []entry{{role: "tool", text: "history cleared - starting a new session"}}
					return m, nil
				}
				m.entries = append(m.entries, entry{role: "you", text: input})
				m.messages = append(m.messages, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))
				m.state = stateWaiting
				m.errMsg = ""
				m.streamBuf = ""
				m.iterCh = make(chan agent.IterationMsg, 10)
				m.streamCh = make(chan string, 100)
				return m, tea.Batch(
					agent.CallAPI(m.client, m.messages, model, m.iterCh, m.streamCh),
					m.spin.Tick,
					waitForIteration(m.iterCh),
					waitForChunk(m.streamCh),
				)
			}
		}

	case agent.IterationMsg:
		m.entries = append(m.entries, entry{
			role: "meta",
			text: fmt.Sprintf("[stop: %s | in: %d | out: %d]", msg.StopReason, msg.InputTokens, msg.OutputTokens),
		})
		return m, waitForIteration(m.iterCh)

	case iterDoneMsg:
		return m, nil

	case streamChunkMsg:
		m.streamBuf += string(msg)
		return m, waitForChunk(m.streamCh)

	case streamDoneMsg:
		return m, nil

	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spin, spinCmd = m.spin.Update(msg)
		return m, spinCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.input.Width = msg.Width - 4
		m.logoStr = renderLogo(msg.Width)

	case agent.ResponseMsg:
		m.state = stateCode
		m.streamBuf = ""
		if msg.Err != nil {
			m.messages = m.messages[:len(m.messages)-1]
			m.entries = m.entries[:len(m.entries)-1]
			m.errMsg = fmt.Sprintf("error: %v", msg.Err)
			return m, nil
		}
		for _, tc := range msg.ToolCalls {
			m.entries = append(m.entries, entry{role: "tool", text: tc.Display})
			if tc.Output != "" {
				role := "exec-ok"
				if tc.IsError {
					role = "exec-err"
				}
				m.entries = append(m.entries, entry{role: role, text: tc.Output})
			}
		}
		m.entries = append(m.entries, entry{role: "codeaid", text: msg.Reply})
		m.messages = append(m.messages, anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Reply)))
		saveHistory(m.messages)
		m.entries = append(m.entries, entry{
			role: "meta",
			text: fmt.Sprintf("[stop: %s | model: %s | total in: %d | total out: %d | total: %d]",
				msg.StopReason, msg.ModelUsed, msg.InputTokens, msg.OutputTokens, msg.InputTokens+msg.OutputTokens),
		})
	}

	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m tuiModel) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // Code
		history := loadHistory()
		m.messages = history
		m.historyCount = len(history)
		m.entries = []entry{}
		m.errMsg = ""
		m.state = stateCode
		m.input.Placeholder = "Ask anything...  (exit: menu  clear: screen  clear history: new session)"
		m.input.Focus()
		return m, textinput.Blink
	case 1: // Load History — placeholder
		m.errMsg = "Load History: coming soon"
		return m, nil
	case 2: // Model
		m.modelCursor = currentModelIndex()
		m.state = stateModelSelect
		return m, nil
	case 3: // Exit
		return m, tea.Quit
	}
	return m, nil
}

func currentModelIndex() int {
	for i, m := range availableModels {
		if m.id == model {
			return i
		}
	}
	return 0
}

func (m tuiModel) View() string {
	var b strings.Builder
	sep := strings.Repeat("─", max(min(m.width, 100), 40))

	b.WriteString("codeaid\n")
	b.WriteString(sep + "\n\n")

	switch m.state {
	case stateAPIKey:
		b.WriteString("No API key found. Enter your Anthropic API key:\n\n")
		b.WriteString("> " + m.input.View() + "\n")

	case stateMenu:
		if m.logoStr != "" {
			b.WriteString(m.logoStr + "\n")
		}
		for i, item := range menuItems {
			cursor := "  "
			if i == m.cursor {
				cursor = "▶ "
			}
			b.WriteString(fmt.Sprintf("%s%-20s%s\n", cursor, item.label, item.desc))
		}
		b.WriteString("\n↑/↓ navigate   enter select   1-4 shortcut   ctrl+c quit\n")
		if m.errMsg != "" {
			b.WriteString("\n" + m.errMsg + "\n")
		}

	case stateModelSelect:
		b.WriteString(fmt.Sprintf("Select a model  (active: %s)\n\n", model))
		for i, mdl := range availableModels {
			cursor := "  "
			if i == m.modelCursor {
				cursor = "▶ "
			}
			active := ""
			if mdl.id == model {
				active = "  ✓"
			}
			b.WriteString(fmt.Sprintf("%s%-32s%s%s\n", cursor, mdl.id, mdl.desc, active))
		}
		b.WriteString("\n↑/↓ navigate   enter select   esc back\n")

	case stateCode, stateWaiting:
		if m.historyCount > 0 {
			b.WriteString(fmt.Sprintf("%d messages loaded from history  (type 'clear' to reset)\n\n", m.historyCount))
		}
		for _, e := range m.entries {
			switch e.role {
			case "you":
				b.WriteString("you: " + e.text + "\n\n")
			case "tool":
				b.WriteString(styles.BlockToolCallStyle.Render("✓ "+e.text) + "\n\n")
			case "exec-ok":
				b.WriteString(styles.ExecOkStyle.Render(e.text) + "\n\n")
			case "exec-err":
				b.WriteString(styles.ExecErrStyle.Render(e.text) + "\n\n")
			case "codeaid":
				b.WriteString("codeaid: " + e.text + "\n\n")
			case "meta":
				b.WriteString(e.text + "\n\n")
			}
		}
		if m.state == stateWaiting {
			if m.streamBuf != "" {
				b.WriteString("codeaid: " + m.streamBuf + "\n\n")
			} else {
				b.WriteString("codeaid: " + m.spin.View() + " thinking...\n\n")
			}
		}
		if m.errMsg != "" {
			b.WriteString(m.errMsg + "\n\n")
		}
		b.WriteString(sep + "\n")
		b.WriteString("> " + m.input.View())
	}

	return b.String()
}

func runTUI() {
	p := tea.NewProgram(newTUI())
	if _, err := p.Run(); err != nil {
		log.Fatalf("TUI error: %v", err)
	}
}
