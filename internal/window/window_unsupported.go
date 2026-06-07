//go:build !windows && !linux

package window

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

var ErrUnsupported = errors.New("wtrans is only supported on Windows")

func ListVisible() ([]Window, error) {
	return nil, ErrUnsupported
}

func SetOpacity(_ Window, _ int) error {
	return ErrUnsupported
}

func Restore(_ Window) error {
	return ErrUnsupported
}

func Diagnose(w io.Writer) error {
	fmt.Fprintf(w, "backend: unsupported\n")
	fmt.Fprintf(w, "platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	return nil
}

func InstallGNOMEExtension(_ io.Writer) error {
	return ErrUnsupported
}

func GNOMEExtensionStatus(_ io.Writer) error {
	return ErrUnsupported
}
