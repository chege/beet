package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	configDir, err := prepareConfig()
	if err != nil {
		log.Fatalf("prepare config: %v", err)
	}

	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "templates":
		names, err := listTemplates(configDir)
		if err != nil {
			log.Fatalf("list templates: %v", err)
		}
		for _, name := range names {
			fmt.Println(name)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", args[0])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: pf [input] | pf -t <template> | pf templates | pf doctor")
}
