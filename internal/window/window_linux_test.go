//go:build linux

package window

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

type fakeRunner struct {
	outputs map[string]string
	paths   map[string]bool
	calls   []string
}

func (f *fakeRunner) Run(name string, args ...string) (string, error) {
	call := name + " " + strings.Join(args, " ")
	f.calls = append(f.calls, call)
	if output, ok := f.outputs[call]; ok {
		return output, nil
	}

	return "", errors.New("unexpected command: " + call)
}

func (f *fakeRunner) LookPath(file string) (string, error) {
	if f.paths[file] {
		return "/usr/bin/" + file, nil
	}

	return "", errors.New("not found")
}

func TestDetectBackend(t *testing.T) {
	introspectCall := "gdbus call --session --dest org.gnome.Shell --object-path /org/gnome/Shell/Extensions/WTrans --method org.freedesktop.DBus.Introspectable.Introspect"
	tests := []struct {
		name string
		env  map[string]string
		r    *fakeRunner
		want string
	}{
		{name: "hyprland env", env: map[string]string{"HYPRLAND_INSTANCE_SIGNATURE": "abc"}, want: backendHyprland},
		{name: "sway env", env: map[string]string{"SWAYSOCK": "/run/sway.sock"}, want: backendSway},
		{name: "x11 env", env: map[string]string{"XDG_SESSION_TYPE": "x11"}, want: backendX11},
		{name: "display env", env: map[string]string{"DISPLAY": ":0"}, want: backendX11},
		{name: "unsupported wayland ignores xwayland display", env: map[string]string{"XDG_SESSION_TYPE": "wayland", "DISPLAY": ":0"}, want: ""},
		{name: "x11 override on wayland", env: map[string]string{"XDG_SESSION_TYPE": "wayland", "DISPLAY": ":0", "WTRANS_BACKEND": "x11"}, want: backendX11},
		{name: "gnome override", env: map[string]string{"XDG_SESSION_TYPE": "wayland", "WTRANS_BACKEND": "gnome"}, want: backendGNOME},
		{
			name: "gnome extension available",
			env:  map[string]string{"XDG_CURRENT_DESKTOP": "ubuntu:GNOME", "XDG_SESSION_TYPE": "wayland"},
			r: &fakeRunner{
				paths:   map[string]bool{"gdbus": true},
				outputs: map[string]string{introspectCall: "('<node><interface name=\"org.gnome.Shell.Extensions.WTrans\"/></node>',)"},
			},
			want: backendGNOME,
		},
		{
			name: "gnome extension object without bridge interface is unsupported",
			env:  map[string]string{"XDG_CURRENT_DESKTOP": "ubuntu:GNOME", "XDG_SESSION_TYPE": "wayland"},
			r: &fakeRunner{
				paths:   map[string]bool{"gdbus": true},
				outputs: map[string]string{introspectCall: "('<node/>',)"},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := tt.r
			if r == nil {
				r = &fakeRunner{}
			}
			got := detectBackend(r, func(key string) string {
				return tt.env[key]
			})
			if got != tt.want {
				t.Fatalf("detectBackend = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAlphaHex(t *testing.T) {
	tests := map[int]string{
		100: "0xffffffff",
		70:  "0xb3333332",
		50:  "0x7fffffff",
		20:  "0x33333333",
	}

	for input, want := range tests {
		if got := alphaHex(input); got != want {
			t.Fatalf("alphaHex(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestParseWMCtrlLine(t *testing.T) {
	win, ok := parseWMCtrlLine("0x03a00007  0 4242 host Visual Studio Code")
	if !ok {
		t.Fatal("parseWMCtrlLine returned ok=false")
	}

	if win.ID != "0x03a00007" || win.PID != 4242 || win.Title != "Visual Studio Code" {
		t.Fatalf("parsed window = %#v", win)
	}
}

func TestParseWMCtrlLineAllowsUnknownPID(t *testing.T) {
	win, ok := parseWMCtrlLine("0x03a00007  0 -1 host Utility")
	if !ok {
		t.Fatal("parseWMCtrlLine returned ok=false")
	}

	if win.PID != 0 || win.Title != "Utility" {
		t.Fatalf("parsed window = %#v", win)
	}
}

func TestParseWMClass(t *testing.T) {
	got := parseWMClass(`WM_CLASS(STRING) = "code", "Code"`)
	if got != "Code" {
		t.Fatalf("parseWMClass = %q, want Code", got)
	}
}

func TestSetX11OpacityCommand(t *testing.T) {
	r := &fakeRunner{
		paths: map[string]bool{"xprop": true},
		outputs: map[string]string{
			"xprop -id 0xabc -f _NET_WM_WINDOW_OPACITY 32c -set _NET_WM_WINDOW_OPACITY 0x7fffffff": "",
		},
	}

	err := setX11Opacity(r, Window{ID: "0xabc", Backend: backendX11}, 50)
	if err != nil {
		t.Fatalf("setX11Opacity returned error: %v", err)
	}
}

func TestSetSwayOpacityCommand(t *testing.T) {
	r := &fakeRunner{
		paths: map[string]bool{"swaymsg": true},
		outputs: map[string]string{
			"swaymsg [con_id=42] opacity set 0.7": "",
		},
	}

	err := setSwayOpacity(r, Window{ID: "42", Backend: backendSway}, 70)
	if err != nil {
		t.Fatalf("setSwayOpacity returned error: %v", err)
	}
}

func TestSetHyprlandOpacityCommand(t *testing.T) {
	r := &fakeRunner{
		paths: map[string]bool{"hyprctl": true},
		outputs: map[string]string{
			"hyprctl setprop address:0xabc alpha 0.85":         "",
			"hyprctl setprop address:0xabc alphainactive 0.85": "",
		},
	}

	err := setHyprlandOpacity(r, Window{ID: "0xabc", Backend: backendHyprland}, 85)
	if err != nil {
		t.Fatalf("setHyprlandOpacity returned error: %v", err)
	}
}

func TestParseGDBusString(t *testing.T) {
	got, err := parseGDBusString(`('[{"id":"1","title":"Bob\'s Editor"}]',)`)
	if err != nil {
		t.Fatalf("parseGDBusString returned error: %v", err)
	}

	want := `[{"id":"1","title":"Bob's Editor"}]`
	if got != want {
		t.Fatalf("parseGDBusString = %q, want %q", got, want)
	}
}

func TestParseGNOMEShellMajor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
		ok    bool
	}{
		{name: "gnome 44", input: "GNOME Shell 44.9", want: 44, ok: true},
		{name: "gnome 45", input: "GNOME Shell 45.2", want: 45, ok: true},
		{name: "ubuntu suffix", input: "GNOME Shell 46.0-ubuntu", want: 46, ok: true},
		{name: "unparseable", input: "GNOME Shell", ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseGNOMEShellMajor(tt.input)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("parseGNOMEShellMajor(%q) = %d, %v; want %d, %v", tt.input, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestGNOMEExtensionMetadataForMajor(t *testing.T) {
	legacy := gnomeExtensionMetadataForMajor(44)
	if !strings.Contains(legacy, `"44"`) || strings.Contains(legacy, `"45"`) {
		t.Fatalf("legacy metadata shell versions look wrong: %s", legacy)
	}

	modern := gnomeExtensionMetadataForMajor(45)
	if !strings.Contains(modern, `"45"`) || strings.Contains(modern, `"44"`) {
		t.Fatalf("modern metadata shell versions look wrong: %s", modern)
	}

	future := gnomeExtensionMetadataForMajor(51)
	if !strings.Contains(future, `"51"`) {
		t.Fatalf("future metadata does not include detected shell version: %s", future)
	}
}

func TestListGNOMEWindows(t *testing.T) {
	call := "gdbus call --session --dest org.gnome.Shell --object-path /org/gnome/Shell/Extensions/WTrans --method org.gnome.Shell.Extensions.WTrans.ListWindows"
	r := &fakeRunner{
		paths: map[string]bool{"gdbus": true},
		outputs: map[string]string{
			call: `('[{"id":"11","pid":123,"process":"org.gnome.Nautilus","class":"Nautilus","title":"Files"}]',)`,
		},
	}

	windows, err := listGNOMEWindows(r)
	if err != nil {
		t.Fatalf("listGNOMEWindows returned error: %v", err)
	}
	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	want := Window{ID: "11", PID: 123, Process: "org.gnome.Nautilus", Class: "Nautilus", Title: "Files", Visible: true, Backend: backendGNOME}
	if !reflect.DeepEqual(windows[0], want) {
		t.Fatalf("windows[0] = %#v, want %#v", windows[0], want)
	}
}

func TestSetGNOMEOpacityCommand(t *testing.T) {
	r := &fakeRunner{
		paths: map[string]bool{"gdbus": true},
		outputs: map[string]string{
			"gdbus call --session --dest org.gnome.Shell --object-path /org/gnome/Shell/Extensions/WTrans --method org.gnome.Shell.Extensions.WTrans.SetOpacity 11 70": "",
		},
	}

	err := setGNOMEOpacity(r, Window{ID: "11", Backend: backendGNOME}, 70)
	if err != nil {
		t.Fatalf("setGNOMEOpacity returned error: %v", err)
	}
}

func TestUnsupportedWaylandErrorForTTY(t *testing.T) {
	err := unsupportedWaylandError(func(key string) string {
		if key == "XDG_SESSION_TYPE" {
			return "tty"
		}
		return ""
	})
	if err == nil || !strings.Contains(err.Error(), "outside a graphical desktop session") {
		t.Fatalf("unsupportedWaylandError = %v", err)
	}
}

func TestWalkSway(t *testing.T) {
	root := swayNode{
		Nodes: []swayNode{
			{
				ID:    7,
				PID:   123,
				Name:  "Terminal",
				AppID: "kitty",
				Type:  "con",
			},
		},
	}

	var windows []Window
	walkSway(root, &windows)
	if len(windows) != 1 {
		t.Fatalf("len(windows) = %d, want 1", len(windows))
	}

	want := Window{Handle: 7, ID: "7", PID: 123, Process: "", Class: "kitty", Title: "Terminal", Visible: true, Backend: backendSway}
	if !reflect.DeepEqual(windows[0], want) {
		t.Fatalf("windows[0] = %#v, want %#v", windows[0], want)
	}
}
