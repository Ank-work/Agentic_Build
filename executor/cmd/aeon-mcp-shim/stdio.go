package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/Ank-work/Agentic_Build/executor/internal/mcpshim"
)

func serveStdio() error {
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		out, err := mcpshim.HandleLine(line)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(os.Stdout, "%s\n", out); err != nil {
			return err
		}
	}
	return sc.Err()
}
