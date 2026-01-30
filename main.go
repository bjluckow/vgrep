package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

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
	dualMode   bool
}

var matchStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("1")).
	Foreground(lipgloss.Color("15"))

var dimStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

var gutterStyle = dimStyle

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

		m.input.Width = msg.Width - 10

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
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

	var content string

	if !m.dualMode {
		// ===== single column
		var numbered []string
		width := 1 + len(strconv.Itoa(len(m.lines)))

		for i, line := range m.rendered {
			gutter := fmt.Sprintf("%*d | ", width, i+1)
			gutter = gutterStyle.Render(gutter)
			numbered = append(numbered, gutter+" "+line)
		}

		content = strings.Join(numbered, "\n")

	} else {
		sep := " │ "
		sepWidth := lipgloss.Width(sep)

		// exact 50/50 columns
		colWidth := (m.width - sepWidth) / 2

		leftStyle := lipgloss.NewStyle().Width(colWidth)
		rightStyle := lipgloss.NewStyle().Width(colWidth)

		var left []string
		var right []string

		for i, line := range m.rendered {
			if m.matches[i] {
				right = append(right, rightStyle.Render(line))
			} else {
				left = append(left, leftStyle.Render(line))
			}
		}

		max := len(left)
		if len(right) > max {
			max = len(right)
		}

		var rows []string

		for i := 0; i < max; i++ {
			l := strings.Repeat(" ", colWidth)
			r := strings.Repeat(" ", colWidth)

			if i < len(left) {
				l = left[i]
			}
			if i < len(right) {
				r = right[i]
			}

			rows = append(rows, l+sep+r)
		}

		content = strings.Join(rows, "\n")
		m.view.SetContent(content)
	}

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

type flags struct {
	dualMode bool
}

func parseFlags(vgrepFlags []string) *flags {
	var result flags
	fs := flag.NewFlagSet("vgrep", flag.ExitOnError)
	fs.BoolVar(&result.dualMode, "d", false, "dual column mode")

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

	fmt.Println()
	for i, line := range result.lines {
		if result.matches[i] {
			fmt.Println(line)
		}
	}
}
