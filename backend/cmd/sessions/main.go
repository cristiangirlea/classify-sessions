// Command sessions classifies Claude Code session logs and reports spend.
// Only `version` exists so far; serve, summary, evaluate, prompts and crosscheck
// (all requiring --input DIR) are added with the packages they call.
package main

import (
	"fmt"
	"os"
	"runtime"
)

const version = "0.0.0"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("sessions %s (%s)\n", version, runtime.Version())
		return
	}
	fmt.Fprintln(os.Stderr, "usage: sessions version")
	os.Exit(2)
}
