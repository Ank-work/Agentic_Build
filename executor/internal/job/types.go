package job

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Argv is a process argument vector. Never a shell string (§6.1, §18).
type Argv []string

// UnmarshalJSON rejects a single string (shell command) and other non-arrays.
func (a *Argv) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || data[0] != '[' {
		return fmt.Errorf("argv must be an array, never a shell string")
	}
	var strs []string
	if err := json.Unmarshal(data, &strs); err != nil {
		return fmt.Errorf("argv must be an array of strings")
	}
	*a = Argv(strs)
	return nil
}

// Job is the §6.1 envelope (apiVersion/kind/metadata/spec).
type Job struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
}

// Metadata identifies a job. Extra keys are rejected at parse time.
type Metadata struct {
	ID        string            `json:"id"`
	CreatedAt string            `json:"createdAt"`
	Workflow  string            `json:"workflow,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// Spec is the job body. additionalProperties are rejected at parse time.
type Spec struct {
	Stage           string            `json:"stage"`
	Argv            Argv              `json:"argv"`
	Cwd             string            `json:"cwd,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	TimeoutSeconds  int               `json:"timeoutSeconds"`
	Network         *Network          `json:"network,omitempty"`
	CapabilityToken string            `json:"capabilityToken,omitempty"`
	Artifacts       *Artifacts        `json:"artifacts,omitempty"`
}

// Network controls egress. Default mode is none (§6.1, §12.3).
type Network struct {
	Mode         string        `json:"mode"`
	Destinations []Destination `json:"destinations,omitempty"`
}

// Destination is a single allowlisted host (and optional port).
type Destination struct {
	Host string `json:"host"`
	Port int    `json:"port,omitempty"`
}

// Artifacts names paths for PC→R2 upload after exit.
type Artifacts struct {
	Paths  []string `json:"paths,omitempty"`
	Bucket string   `json:"bucket,omitempty"`
	Prefix string   `json:"prefix,omitempty"`
}

// DestinationHosts returns host names from spec.network.destinations.
func (j *Job) DestinationHosts() []string {
	if j == nil || j.Spec.Network == nil {
		return nil
	}
	out := make([]string, 0, len(j.Spec.Network.Destinations))
	for _, d := range j.Spec.Network.Destinations {
		if d.Host != "" {
			out = append(out, d.Host)
		}
	}
	return out
}

// NetworkMode returns the mode after defaults (empty if job is nil).
func (j *Job) NetworkMode() string {
	if j == nil || j.Spec.Network == nil {
		return ""
	}
	return j.Spec.Network.Mode
}
