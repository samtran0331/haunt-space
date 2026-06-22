package main

import (
	"fmt"
	"os"
)

const usage = `hsp — haunt-space layout engine

Usage:
  hsp boo                    Create a new template (interactive wizard)
  hsp summon [template]      Launch a saved template, or browse if no name given
  hsp <template>             Launch template directly (shorthand)
`

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "boo":
		if err := runBoo(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

	case "summon":
		if len(os.Args) < 3 {
			if err := runWizard(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := launchTemplate(os.Args[2]); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}

	default:
		// Treat the argument as a template name shorthand.
		name := os.Args[1]
		if _, err := os.Stat(templatePath(name)); err == nil {
			if err := launchTemplate(name); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := runNotFound(name); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		}
	}
}
