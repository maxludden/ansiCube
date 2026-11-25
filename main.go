// Command ansicube is an interactive ANSI 256-color explorer for the
// terminal. Click any tile to copy its color number to the clipboard.
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
)

// version is set at build time by GoReleaser via -ldflags "-X main.version=…".
var version = "dev"

// buildVersion returns the injected version, or the module version recorded
// by `go install …@vX.Y.Z` when no version was injected.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\nInteractive ANSI 256-color explorer.\n\nFlags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("ansicube", buildVersion())
		return
	}

	if _, err := tea.NewProgram(newModel(clipboard.WriteAll)).Run(); err != nil {
		fmt.Fprintf(os.Stderr, "ansicube: %v\n", err)
		os.Exit(1)
	}
}
