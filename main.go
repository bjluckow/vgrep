package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

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
	finalPattern string
}

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

func splitArgs() ([]string, []string) {
	args := os.Args[1:]
	for i, a := range args {
		if a == "--" {
			return args[:i], args[i+1:]
		}
	}
	return args, nil
}

func writeOutput(m model, f *flags) {
	if f.patternOut {
		fmt.Println(m.finalPattern)
		return
	}

	for i, line := range m.lines {
		if m.matches[i] {
			fmt.Println(line)
		}
	}
}

type flags struct {
	dualMode   bool
	patternOut bool
}

func parseFlags(vgrepFlags []string) *flags {
	var result flags
	fs := flag.NewFlagSet("vgrep", flag.ExitOnError)
	fs.BoolVar(&result.dualMode, "d", false, "dual column mode (unmatched | matched)")
	fs.BoolVar(&result.patternOut, "p", false, "output regexp instead of matches")

	fs.Parse(vgrepFlags)
	return &result
}

func main() {
	vgrepFlags, grepArgs := splitArgs()
	parsedFlags := parseFlags(vgrepFlags)

	m := initialModel(grepArgs)
	m.dualMode = parsedFlags.dualMode

	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		panic(err)
	}

	result := finalModel.(model)

	// user cancelled
	if result.matches == nil {
		return
	}

	writeOutput(result, parsedFlags)
}
