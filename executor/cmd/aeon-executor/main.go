package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Ank-work/Agentic_Build/executor/internal/allowlist"
	"github.com/Ank-work/Agentic_Build/executor/internal/capability"
	"github.com/Ank-work/Agentic_Build/executor/internal/gitqueue"
	"github.com/Ank-work/Agentic_Build/executor/internal/job"
)

func main() {
	allowlistPath := flag.String("allowlist", "", "argv[0] allowlist file (default: embedded default.txt)")
	jobsInbox := flag.String("jobs-inbox", "", "fallback git-queue inbox (not the happy path)")
	validateFile := flag.String("validate-file", "", "load and validate a job JSON file, then exit")
	flag.Parse()

	allowed := allowlist.Default()
	if *allowlistPath != "" {
		body, err := os.ReadFile(*allowlistPath)
		if err != nil {
			fail(fmt.Errorf("allowlist: %w", err))
		}
		allowed = allowlist.ParseList(string(body))
	}

	switch {
	case *validateFile != "":
		if err := validatePath(*validateFile, allowed); err != nil {
			fail(err)
		}
		return
	case *jobsInbox != "":
		outbox := gitqueue.OutboxDir(*jobsInbox)
		if err := gitqueue.ProcessInbox(*jobsInbox, outbox); err != nil {
			fail(err)
		}
		fmt.Println("ok")
		fmt.Println("gitqueue=fallback-only")
		return
	default:
		fmt.Println("aeon-executor skeleton (not a science executor)")
		fmt.Println("flags: -validate-file, -allowlist, -jobs-inbox")
	}
}

func validatePath(path string, allowed allowlist.Set) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	j, err := job.Parse(raw)
	if err != nil {
		return err
	}
	if err := allowlist.Check([]string(j.Spec.Argv), allowed); err != nil {
		return err
	}
	if err := capability.Validate(j.Spec.Stage, j.Spec.CapabilityToken, j.Spec.Argv[0], j.DestinationHosts()); err != nil {
		return err
	}
	fmt.Println("ok")
	fmt.Printf("id=%s\n", j.Metadata.ID)
	fmt.Printf("stage=%s\n", j.Spec.Stage)
	fmt.Printf("network.mode=%s\n", j.NetworkMode())
	return nil
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", err)
	os.Exit(1)
}
