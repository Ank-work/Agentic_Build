package job

// ApplyDefaults implements the §6.1 executor invariant: missing network
// or missing/empty network.mode is treated as { "mode": "none" }.
// This is not a schema relaxation — apply defaults, then Validate.
func ApplyDefaults(j *Job) {
	if j == nil {
		return
	}
	if j.Spec.Network == nil {
		j.Spec.Network = &Network{Mode: "none"}
		return
	}
	if j.Spec.Network.Mode == "" {
		j.Spec.Network.Mode = "none"
	}
}
