package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bluegardenproject/github-butler/cmd"
)

// Version is the binary version, set at build time via:
//
//	-ldflags "-X main.Version=v1.2.3"
//
// Release Please bumps this on every release via the `extra-files`
// entry in release-please-config.json, so the in-tree default also
// matches the latest tagged release between rebuilds.
var Version = "0.3.0" // x-release-please-version

// BuildTime is the UTC timestamp the binary was built at, set via:
//
//	-ldflags "-X main.BuildTime=2026-05-06T17:00:00Z"
var BuildTime = "unknown"

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cmd.SetVersion(Version, BuildTime)

	if err := cmd.Run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
