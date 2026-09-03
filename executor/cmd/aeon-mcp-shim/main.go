package main

import (
	"fmt"
	"os"

	"github.com/Ank-work/Agentic_Build/executor/internal/mcpshim"
)

func main() {
	listTools := false
	probe := ""
	stdio := false
	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-list-tools", "--list-tools":
			listTools = true
		case "-stdio", "--stdio":
			stdio = true
		case "-probe-tool", "--probe-tool":
			if i+1 >= len(os.Args) {
				fmt.Fprintln(os.Stderr, "error: -probe-tool requires a name")
				os.Exit(2)
			}
			i++
			probe = os.Args[i]
		default:
			fmt.Fprintf(os.Stderr, "unknown flag %s\n", os.Args[i])
			os.Exit(2)
		}
	}

	if probe != "" {
		res := mcpshim.CallTool(probe)
		fmt.Println(res.Message)
		if res.Denied {
			os.Exit(0)
		}
		os.Exit(0)
	}

	if stdio {
		if err := serveStdio(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Default and -list-tools: print advertised tools only.
	_ = listTools
	for _, t := range mcpshim.ListTools() {
		fmt.Println(t.Name)
	}
}
