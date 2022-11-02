package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	osc = "\u001B]"
	bel = "\u0007"
)

func url(url, text string) string {
	return osc + "8;;" + url + bel + text + osc + "8;;" + bel
}

func proto(scheme, path string) string {
	return scheme + "://" + path
}

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
	relPaths, err := os.ReadDir(".")
	if err != nil {
		return fmt.Errorf("read local dir: %w", err)
	}
	for _, path := range relPaths {
		matchPrefixes = append(matchPrefixes, regexp.QuoteMeta(path.Name()))
	}

	// add ~/ home dir prefix
	matchPrefixes = append(matchPrefixes, regexp.QuoteMeta("~"))

	expr, err := regexp.Compile(fmt.Sprintf(`(?:%s)(?:/[^$\s\;~\:]+)?`, strings.Join(matchPrefixes, "|")))
	if err != nil {
		return fmt.Errorf("compile path regexp: %w", err)
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		replaced := expr.ReplaceAllStringFunc(scanner.Text(), func(match string) string {
			if strings.HasPrefix(match, "~/") {
				match = filepath.Join(home, match[2:])
			}
			abs, err := filepath.Abs(string(match))
			if err != nil {
				return match
			}
			abs = filepath.Join(hostname, abs)
			return url(proto("file", abs), match)
		})
		os.Stdout.Write([]byte(replaced))
		os.Stdout.Write([]byte{'\n'})
	}

	return nil
}
