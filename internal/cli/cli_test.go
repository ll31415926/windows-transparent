package cli

import "testing"

func TestParseSetCommand(t *testing.T) {
	cmd, err := Parse([]string{"set", "--process", "notepad.exe", "--opacity", "70"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandSet {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandSet)
	}
	if cmd.Process != "notepad.exe" {
		t.Fatalf("cmd.Process = %q, want notepad.exe", cmd.Process)
	}
	if cmd.Opacity != 70 {
		t.Fatalf("cmd.Opacity = %d, want 70", cmd.Opacity)
	}
}

func TestParseListCommandWithProcess(t *testing.T) {
	cmd, err := Parse([]string{"list", "--process=Code.exe"})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandList {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandList)
	}
	if cmd.Process != "Code.exe" {
		t.Fatalf("cmd.Process = %q, want Code.exe", cmd.Process)
	}
}

func TestParseApplyCommandWithConfig(t *testing.T) {
	cmd, err := Parse([]string{"apply", "--config", `C:\tmp\wtrans.json`})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if cmd.Name != CommandApply {
		t.Fatalf("cmd.Name = %q, want %q", cmd.Name, CommandApply)
	}
	if cmd.ConfigPath != `C:\tmp\wtrans.json` {
		t.Fatalf("cmd.ConfigPath = %q", cmd.ConfigPath)
	}
}

func TestParseSetCommandRequiresProcess(t *testing.T) {
	_, err := Parse([]string{"set", "--opacity", "70"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseSetCommandRequiresOpacity(t *testing.T) {
	_, err := Parse([]string{"set", "--process", "notepad.exe"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseSetCommandRejectsBadOpacity(t *testing.T) {
	_, err := Parse([]string{"set", "--process", "notepad.exe", "--opacity", "10"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}

func TestParseRestoreCommandRequiresProcess(t *testing.T) {
	_, err := Parse([]string{"restore"})
	if err == nil {
		t.Fatal("Parse returned nil error, want usage error")
	}
	if !IsUsageError(err) {
		t.Fatalf("Parse error = %T, want UsageError", err)
	}
}
