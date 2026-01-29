package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	lines      []string
	matches    []bool
	rendered   []string
	input      textinput.Model
	view       viewport.Model
	width      int
	height     int
	matchCount int
	regexErr   error
}

var matchStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("1")).
	Foreground(lipgloss.Color("15"))

var dimStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

func loadInput() []string {
	info, _ := os.Stdin.Stat()

	var reader io.Reader

	// stdin piped
	if info.Mode()&os.ModeCharDevice == 0 {
		reader = os.Stdin
	} else if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		reader = file
	} else {
		return []string{"(no input — pipe data or pass a file)"}
	}

	var lines []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

func initialModel() model {
	lines := loadInput()

	ti := textinput.New()
	ti.Placeholder = "regex"
	ti.Focus()

	vp := viewport.New(0, 0)

	m := model{
		lines:   lines,
		matches: make([]bool, len(lines)),
		input:   ti,
		view:    vp,
	}

	m.applyRegex("")
	return m
}

func (m *model) applyRegex(pattern string) {
	if pattern == "" {
		m.rendered = m.lines
		for i := range m.matches {
			m.matches[i] = false
		}
		m.matchCount = 0
		m.regexErr = nil
		return
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		m.regexErr = err
		return
	}

	m.matchCount = 0
	var out []string

	for idx, line := range m.lines {
		if re.MatchString(line) {
			m.matches[idx] = true
			m.matchCount++
			highlighted := re.ReplaceAllStringFunc(line, func(s string) string {
				return matchStyle.Render(s)
			})
			out = append(out, highlighted)
		} else {
			m.matches[idx] = false
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

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, tea.Quit
		case "ctrl+c", "esc":
			os.Exit(1)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	m.applyRegex(m.input.Value())

	content := strings.Join(m.rendered, "\n")
	m.view.SetContent(content)

	return m, cmd
}

func (m model) View() string {
	status := fmt.Sprintf("(%d/%d matches)", m.matchCount, len(m.lines))

	if m.regexErr != nil {
		status = fmt.Sprintf("regex error: %v", m.regexErr)
	}

	statusLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Render(status)

	return m.view.View() + "\n" + statusLine + " " + m.input.View()
}

func main() {
	m := initialModel()
	p := tea.NewProgram(m, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		panic(err)
	}

	result := finalModel.(model)
	for i, line := range result.lines {
		if result.matches[i] {
			fmt.Println(line)
		}
	}
}
