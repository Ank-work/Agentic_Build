package job

import (
	"bytes"
	"encoding/json"
	"fmt"
)

const (
	APIVersion = "aeon.dev/v1"
	KindJob    = "Job"
	StageReader   = "reader"
	StageExecutor = "executor"
	NetworkNone      = "none"
	NetworkAllowlist = "allowlist"
)

// Parse decodes a job document with unknown-key rejection, applies
// defaults, then Validate. Capability tokens are never logged.
func Parse(data []byte) (*Job, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var j Job
	if err := dec.Decode(&j); err != nil {
		return nil, fmt.Errorf("job parse: %w", err)
	}
	ApplyDefaults(&j)
	if err := Validate(&j); err != nil {
		return nil, err
	}
	return &j, nil
}

// Validate enforces §6.1 invariants after defaults.
// Do not include capabilityToken in any error string.
func Validate(j *Job) error {
	if j == nil {
		return fmt.Errorf("job is nil")
	}
	if j.APIVersion != APIVersion {
		return fmt.Errorf("apiVersion must be %s", APIVersion)
	}
	if j.Kind != KindJob {
		return fmt.Errorf("kind must be %s", KindJob)
	}
	if j.Metadata.ID == "" {
		return fmt.Errorf("metadata.id is required")
	}
	if j.Metadata.CreatedAt == "" {
		return fmt.Errorf("metadata.createdAt is required")
	}
	if err := validateSpec(&j.Spec); err != nil {
		return err
	}
	return nil
}

func validateSpec(s *Spec) error {
	switch s.Stage {
	case StageReader, StageExecutor:
	default:
		return fmt.Errorf("spec.stage must be reader or executor")
	}
	if len(s.Argv) == 0 {
		return fmt.Errorf("spec.argv must be a non-empty array, never a shell string")
	}
	for i, a := range s.Argv {
		if a == "" {
			return fmt.Errorf("spec.argv[%d] must be a non-empty string", i)
		}
	}
	if s.TimeoutSeconds < 1 || s.TimeoutSeconds > 86400 {
		return fmt.Errorf("spec.timeoutSeconds must be in 1..86400")
	}
	if s.Network == nil {
		return fmt.Errorf("spec.network is required after defaults")
	}
	switch s.Network.Mode {
	case NetworkNone:
	case NetworkAllowlist:
		if len(s.Network.Destinations) == 0 {
			return fmt.Errorf("network.mode allowlist requires destinations[]")
		}
		for i, d := range s.Network.Destinations {
			if d.Host == "" {
				return fmt.Errorf("network.destinations[%d].host is required", i)
			}
			if d.Port != 0 && (d.Port < 1 || d.Port > 65535) {
				return fmt.Errorf("network.destinations[%d].port out of range", i)
			}
		}
	default:
		return fmt.Errorf("network.mode must be none or allowlist")
	}
	switch s.Stage {
	case StageReader:
		if s.CapabilityToken != "" {
			return fmt.Errorf("reader must not receive a usable capabilityToken")
		}
	case StageExecutor:
		if s.CapabilityToken == "" {
			return fmt.Errorf("executor must present a capability token")
		}
	}
	return nil
}
