package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var matchStyle = lipgloss.NewStyle().
	Background(lipgloss.Color("1")).
	Foreground(lipgloss.Color("15"))

var dimStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("8"))

var gutterStyle = dimStyle

func renderContent(m model) string {
	if !m.dualMode {
		return renderSingle(m)
	}
	return renderDual(m)
}

func renderSingle(m model) string {
	var numbered []string
	width := 1 + len(strconv.Itoa(len(m.lines)))

	for i, line := range m.rendered {
		gutter := fmt.Sprintf("%*d | ", width, i+1)
		gutter = gutterStyle.Render(gutter)
		numbered = append(numbered, gutter+" "+line)
	}

	return strings.Join(numbered, "\n")
}

func renderDual(m model) string {
	sep := " │ "
	sepWidth := lipgloss.Width(sep)

	w := m.width
	if w < 20 {
		w = 80
	}

	colWidth := (w - sepWidth) / 2
	if colWidth < 1 {
		colWidth = 1
	}

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

	return strings.Join(rows, "\n")
}
