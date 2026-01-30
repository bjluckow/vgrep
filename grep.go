package main

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"
)

// RunGrep executes grep and returns a match map indexed by line number
func RunGrep(pattern string, grepArgs []string) (map[int]bool, error) {
	matches := make(map[int]bool)

	if pattern == "" {
		return matches, nil
	}

	args := append([]string{"-n"}, grepArgs...)
	args = append(args, pattern)

	cmd := exec.Command("grep", args...)

	out, err := cmd.Output()

	// grep returns exit code 1 when no matches found
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			out = exitErr.Stdout
		} else {
			return nil, err
		}
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))

	for scanner.Scan() {
		line := scanner.Text()

		// format: "12:content"
		idx := strings.Index(line, ":")
		if idx == -1 {
			continue
		}

		n, err := strconv.Atoi(line[:idx])
		if err != nil {
			continue
		}

		matches[n-1] = true // grep is 1-indexed
	}

	return matches, nil
}
