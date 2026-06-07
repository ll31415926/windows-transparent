package window

import "testing"

func TestMatchByProcessIsCaseInsensitive(t *testing.T) {
	windows := []Window{
		{Handle: 1, Process: "notepad.exe"},
		{Handle: 2, Process: "Code.exe"},
		{Handle: 3, Process: "explorer.exe"},
	}

	matches := MatchByProcess(windows, "code.exe")
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}

	if matches[0].Handle != 2 {
		t.Fatalf("matched handle = %d, want 2", matches[0].Handle)
	}
}

func TestMatchByProcessMatchesClass(t *testing.T) {
	windows := []Window{
		{Handle: 1, Process: "electron", Class: "Code"},
		{Handle: 2, Process: "kitty", Class: "kitty"},
	}

	matches := MatchByProcess(windows, "code")
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}

	if matches[0].Handle != 1 {
		t.Fatalf("matched handle = %d, want 1", matches[0].Handle)
	}
}

func TestMatchByProcessWithEmptyProcessReturnsCopy(t *testing.T) {
	windows := []Window{{Handle: 1, Process: "notepad.exe"}}

	matches := MatchByProcess(windows, "")
	if len(matches) != 1 {
		t.Fatalf("len(matches) = %d, want 1", len(matches))
	}

	matches[0].Process = "changed.exe"
	if windows[0].Process != "notepad.exe" {
		t.Fatal("MatchByProcess returned a slice sharing backing storage")
	}
}
