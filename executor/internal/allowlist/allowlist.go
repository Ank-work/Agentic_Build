package allowlist

import (
	"bufio"
	"fmt"
	"path/filepath"
	"strings"

	_ "embed"
)

//go:embed default.txt
var DefaultList string

// Set is a configured argv[0] allowlist.
type Set map[string]struct{}

var deniedBins = map[string]struct{}{
	"sh": {}, "bash": {},
	"cmd": {}, "cmd.exe": {},
	"powershell": {}, "powershell.exe": {},
	"pwsh": {}, "pwsh.exe": {},
}

var deniedArgs = map[string]struct{}{
	"-c": {}, "/c": {}, "-command": {}, "-Command": {},
}

// ParseList loads an allowlist file body (comments and blanks skipped).
func ParseList(body string) Set {
	s := make(Set)
	sc := bufio.NewScanner(strings.NewReader(body))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s[line] = struct{}{}
	}
	return s
}

// Default returns the shipped allowlist (gromacs).
func Default() Set {
	return ParseList(DefaultList)
}

// Check enforces argv[0] allowlisting and hard-rejects shells / -c style
// invocation (§12.2). Does not execute anything (no os/exec).
func Check(argv []string, allowed Set) error {
	if len(argv) == 0 {
		return fmt.Errorf("argv must be a non-empty array")
	}
	argv0 := argv[0]
	if deniedBinary(argv0) {
		return fmt.Errorf("argv[0] is a denied shell interpreter")
	}
	for i := 1; i < len(argv); i++ {
		if deniedArg(argv[i]) {
			return fmt.Errorf("argv contains a denied shell/command flag")
		}
	}
	if allowed == nil {
		allowed = Default()
	}
	if !permittedArgv0(argv0, allowed) {
		return fmt.Errorf("argv[0] is not on the allowlist")
	}
	return nil
}

func deniedBinary(argv0 string) bool {
	norm := strings.ToLower(strings.ReplaceAll(argv0, "\\", "/"))
	base := strings.ToLower(filepath.Base(norm))
	if _, ok := deniedBins[base]; ok {
		return true
	}
	if norm == "/bin/sh" || norm == "/bin/bash" || strings.HasSuffix(norm, "/sh") || strings.HasSuffix(norm, "/bash") {
		return true
	}
	return false
}

func deniedArg(arg string) bool {
	if _, ok := deniedArgs[arg]; ok {
		return true
	}
	return strings.EqualFold(arg, "-Command") || strings.EqualFold(arg, "/c") || arg == "-c"
}

func permittedArgv0(argv0 string, allowed Set) bool {
	if _, ok := allowed[argv0]; ok {
		return true
	}
	base := filepath.Base(strings.ReplaceAll(argv0, "\\", "/"))
	_, ok := allowed[base]
	return ok
}
