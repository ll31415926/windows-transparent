package window

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type runner interface {
	Run(name string, args ...string) (string, error)
	LookPath(file string) (string, error)
}

type osRunner struct{}

func (osRunner) Run(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if message == "" {
			message = strings.TrimSpace(stdout.String())
		}
		if message == "" {
			message = err.Error()
		}

		return "", fmt.Errorf("%s: %s", name, message)
	}

	return stdout.String(), nil
}

func (osRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
