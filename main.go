package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	osc = "\u001B]"
	bel = "\u0007"
)

// \x1b excluded to preserve ANSI colour codes in piped input
const pathDelims = "$;~:\"'(){}[],\x1b"

var commonPrefixes = []string{
	"/bin", "/boot", "/dev", "/etc", "/home", "/lib", "/lib64",
	"/lost+found", "/mnt", "/opt", "/proc", "/root", "/run",
	"/sbin", "/srv", "/sys", "/tmp", "/usr", "/var",
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("find user home dir: %w", err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("find working dir: %w", err)
	}
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("find hostname: %w", err)
	}

	var matchPrefixes []string

	// add possible abs path prefixes
	for _, absPath := range commonPrefixes {
		matchPrefixes = append(matchPrefixes, regexp.QuoteMeta(absPath))
	}

	// add relative to current dir prefixes, fill be expanded to abs paths
	relPaths, err := os.ReadDir(cwd)
	if err != nil {
		return fmt.Errorf("read local dir: %w", err)
	}
	for _, path := range relPaths {
		matchPrefixes = append(matchPrefixes, regexp.QuoteMeta(path.Name()))
	}

	// add ~/ home dir prefix
	matchPrefixes = append(matchPrefixes, regexp.QuoteMeta("~"))

	expr, err := regexp.Compile(fmt.Sprintf(`(?:%s)(?:/[^\s%s]*)?`, strings.Join(matchPrefixes, "|"), regexp.QuoteMeta(pathDelims)))
	if err != nil {
		return fmt.Errorf("compile path regexp: %w", err)
	}

	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		linkPaths(out, scanner.Text(), expr, home, cwd, hostname)
		out.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan stdin: %w", err)
	}

	return nil
}

func linkPaths(out *bufio.Writer, line string, expr *regexp.Regexp, home, cwd, hostname string) {
	last := 0
	for _, loc := range expr.FindAllStringIndex(line, -1) {
		start, end := loc[0], loc[1]
		if !standalone(line, start, end) {
			continue
		}

		match := line[start:end]
		abs := match
		switch {
		case strings.HasPrefix(abs, "~"):
			abs = filepath.Join(home, abs[1:])
		case !filepath.IsAbs(abs):
			abs = filepath.Join(cwd, abs)
		}

		out.WriteString(line[last:start])
		out.WriteString(hyperlink("file://"+filepath.Join(hostname, abs), match))
		last = end
	}
	out.WriteString(line[last:])
}

func hyperlink(target, text string) string {
	return osc + "8;;" + target + bel + text + osc + "8;;" + bel
}

func standalone(line string, start, end int) bool {
	if start > 0 {
		if prev, _ := utf8.DecodeLastRuneInString(line[:start]); pathChar(prev) {
			return false
		}
	}
	if end < len(line) {
		if next, _ := utf8.DecodeRuneInString(line[end:]); pathChar(next) {
			return false
		}
	}
	return true
}

func pathChar(r rune) bool {
	return !unicode.IsSpace(r) && !strings.ContainsRune(pathDelims, r)
}
