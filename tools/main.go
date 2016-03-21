package main

import (
	"log"
	"os"
)

const githubRepoURL = "https://github.com/kirillDanshin/go.uik"

func main() {
	if len(os.Args) == 1 {
		printUsage(os.Args[0])
		return
	}
	switch os.Args[1] {
	case "prepareFont":
		if len(os.Args) == 2 {
			printUsage(os.Args[0])
			return
		} else if len(os.Args[2]) <= 1 {
			log.Fatalf("Input file path is empty. It must be more than one symbol")
		}
		if len(os.Args) == 3 {
			printUsage(os.Args[0])
			return
		} else if len(os.Args[3]) <= 1 {
			log.Fatalf("Output file path is empty. It must be more than one symbol")
		}
		prepareFont(os.Args[2], os.Args[3])
	default:
		printUsage(os.Args[0])
	}
}
