package allowlist

import "testing"

func TestDefaultContainsGromacs(t *testing.T) {
	if err := Check([]string{"gromacs"}, Default()); err != nil {
		t.Fatalf("gromacs should be allowlisted: %v", err)
	}
}

func TestDeniedShells(t *testing.T) {
	cases := [][]string{
		{"cmd.exe", "/c", "echo hi"},
		{"powershell", "-Command", "Get-Host"},
		{"/bin/sh", "-c", "echo hi"},
		{"bash", "-c", "true"},
		{"pwsh", "-c", "1"},
		{"C:\\Windows\\System32\\cmd.exe", "/c", "dir"},
	}
	for _, argv := range cases {
		if err := Check(argv, Default()); err == nil {
			t.Errorf("expected reject of %v", argv)
		}
	}
}

func TestDeniedDashCEvenIfBinaryAllowed(t *testing.T) {
	allowed := Set{"gromacs": {}}
	if err := Check([]string{"gromacs", "-c", "evil"}, allowed); err == nil {
		t.Fatal("expected reject of -c")
	}
}

func TestUnknownBinaryReject(t *testing.T) {
	if err := Check([]string{"not-a-binary"}, Default()); err == nil {
		t.Fatal("expected reject")
	}
}
