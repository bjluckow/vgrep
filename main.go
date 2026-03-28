package main

import (
	"bufio"

	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	flag "github.com/spf13/pflag"
)

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
	dualMode     bool
	patternOut   bool
	showLineNums bool
}

func parseFlags(vgrepFlags []string) *flags {
	var result flags
	fs := flag.NewFlagSet("vgrep", flag.ExitOnError)
	fs.BoolVarP(&result.dualMode, "compare", "c", false, "dual column mode (unmatched | matched)")
	fs.BoolVarP(&result.patternOut, "expr", "e", false, "output regexp instead of matches")
	fs.BoolVarP(&result.showLineNums, "lines", "l", true, "display line numbers")

	fs.Parse(vgrepFlags)
	return &result
}

func main() {
	vgrepFlags, grepArgs := splitArgs()
	parsedFlags := parseFlags(vgrepFlags)

	initRenderer()

	m := initialModel(grepArgs)
	m.dualMode = parsedFlags.dualMode
	m.patternOut = parsedFlags.patternOut
	m.showLineNums = parsedFlags.showLineNums

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr), tea.WithInputTTY())
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
