//go:build !windows

package window

import "errors"

var ErrUnsupported = errors.New("wtrans is only supported on Windows")

func ListVisible() ([]Window, error) {
	return nil, ErrUnsupported
}

func SetOpacity(_ uintptr, _ int) error {
	return ErrUnsupported
}

func Restore(_ uintptr) error {
	return ErrUnsupported
}
