package main

import "fmt"

func printUsage(binName string) {
	fmt.Printf(
		`Usage: %s [command]
	commands:
		prepareFont [font path] [output file path]
`,
		binName,
	)
}
