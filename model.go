package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	lines      []string
	matches    []bool
	grepArgs   []string
	rendered   []string
	input      textinput.Model
	view       viewport.Model
	width      int
	height     int
	matchCount int
	regexErr   error

	dualMode     bool
	patternOut   bool
	showLineNums bool

	finalPattern string
}

func initialModel(grepArgs []string) model {
	lines := loadInput()

	ti := textinput.New()
	ti.Placeholder = "^[.*]$"
	ti.Focus()

	vp := viewport.New(0, 0)

	m := model{
		lines:    lines,
		matches:  make([]bool, len(lines)),
		grepArgs: grepArgs,
		input:    ti,
		view:     vp,
	}

	m.applyRegex("")
	return m
}

func (m *model) applyRegex(pattern string) {
	matchMap, err := RunGrep(pattern, m.grepArgs, m.lines)
	if err != nil {
		m.regexErr = err
		return
	}

	m.matchCount = len(matchMap)

	var out []string

	for i, line := range m.lines {
		if matchMap[i] {
			m.matches[i] = true
			out = append(out, matchStyle.Render(line))
		} else {
			m.matches[i] = false
			out = append(out, dimStyle.Render(line))
		}
	}

	m.rendered = out
	m.regexErr = nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.view.Width = msg.Width
		m.view.Height = msg.Height - 2

		m.input.Width = msg.Width - 10

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.finalPattern = m.input.Value()
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.matches = nil // cancel any output
			return m, tea.Quit
		case "up":
			m.view.ScrollUp(1)
			return m, nil

		case "down":
			m.view.ScrollDown(1)
			return m, nil

		case "pgup":
			m.view.HalfPageUp()
			return m, nil

		case "pgdown":
			m.view.HalfPageDown()
			return m, nil

		case "home":
			m.view.GotoTop()
			return m, nil

		case "end":
			m.view.GotoBottom()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	m.applyRegex(m.input.Value())

	content := renderContent(m)
	m.view.SetContent(content)
	return m, cmd
}

func (m model) View() string {
	percent := float64(m.matchCount) / float64(len(m.lines)) * 100
	status := fmt.Sprintf("%6.2f%%", percent)

	if m.regexErr != nil {
		status = fmt.Sprintf("regex error: %v", m.regexErr)
	}

	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render(status)

	return m.view.View() + "\n" + statusLine + " " + m.input.View()
}
