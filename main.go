package main

import (
	_ "embed"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/UnitVectorY-Labs/json2mdplan/internal/app"
)

var Version = "dev" // This will be set by the build system to the release version

func main() {
	// Set the build version from the build info if not set by the build system
	if Version == "dev" || Version == "" {
		if bi, ok := debug.ReadBuildInfo(); ok {
			if bi.Main.Version != "" && bi.Main.Version != "(devel)" {
				Version = bi.Main.Version
			}
		}
	}

	if err := app.Run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
