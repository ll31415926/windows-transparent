package version

import (
	"fmt"
	"io"
	"runtime"
)

var (
	Version = "dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func Print(w io.Writer) {
	fmt.Fprintf(w, "wtrans %s\n", Version)
	fmt.Fprintf(w, "commit: %s\n", Commit)
	fmt.Fprintf(w, "built: %s\n", Date)
	fmt.Fprintf(w, "go: %s\n", runtime.Version())
	fmt.Fprintf(w, "platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
