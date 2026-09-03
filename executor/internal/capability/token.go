package capability

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Token is a non-crypto stub of an executor capability (§6.1, §12.5).
// Never log the raw token string.
type Token struct {
	Raw        string
	Binaries   []string `json:"binaries"`
	Hosts      []string `json:"hosts"`
	structured bool
}

type tokenJSON struct {
	Binaries []string `json:"binaries"`
	Hosts    []string `json:"hosts"`
}

// ParseToken builds a stub Token. JSON claims are optional; a non-JSON
// non-empty string is an opaque placeholder (covers argv[0] only when
// destinations is empty).
func ParseToken(raw string) Token {
	t := Token{Raw: raw}
	var claims tokenJSON
	if err := json.Unmarshal([]byte(raw), &claims); err == nil {
		t.Binaries = claims.Binaries
		t.Hosts = claims.Hosts
		t.structured = true
	}
	return t
}

// Validate implements reader vs executor capability rules.
// Reader must not get a usable token. Executor must present a token whose
// allowlist covers argv[0] and destinations when network.mode is allowlist.
// The raw token is never included in errors.
func Validate(stage, token, argv0 string, destinations []string) error {
	switch stage {
	case "reader":
		if strings.TrimSpace(token) != "" {
			return fmt.Errorf("reader must not receive a usable capabilityToken")
		}
		return nil
	case "executor":
		if strings.TrimSpace(token) == "" {
			return fmt.Errorf("executor must present a capability token")
		}
		t := ParseToken(token)
		if !t.coversArgv0(argv0) {
			return fmt.Errorf("capability token does not cover argv[0]")
		}
		if len(destinations) > 0 && !t.coversDestinations(destinations) {
			return fmt.Errorf("capability token does not cover destinations")
		}
		return nil
	default:
		return fmt.Errorf("stage must be reader or executor")
	}
}

func (t Token) coversArgv0(argv0 string) bool {
	if strings.TrimSpace(t.Raw) == "" {
		return false
	}
	if !t.structured {
		// Opaque placeholder: argv[0] only, no hosts.
		return argv0 != ""
	}
	return contains(t.Binaries, argv0)
}

func (t Token) coversDestinations(destinations []string) bool {
	if !t.structured {
		return false
	}
	for _, d := range destinations {
		if !contains(t.Hosts, d) {
			return false
		}
	}
	return true
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}
