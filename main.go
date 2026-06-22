package main

import (
	"fmt"
	"os"
)

const usage = `hsp — haunt-space layout engine

Usage:
  hsp wizard                 Launch the interactive template wizard
  hsp launch <template>      Launch a saved template in Ghostty
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "wizard":
		if err := runWizard(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "launch":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "error: launch requires a template name")
			fmt.Fprintln(os.Stderr, "  hsp launch <template>")
			os.Exit(1)
		}
		if err := launchTemplate(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\n", os.Args[1])
		fmt.Print(usage)
		os.Exit(1)
	}
}
