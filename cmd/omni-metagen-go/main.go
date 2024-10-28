package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// These variables are set during build using -ldflags
var (
	buildVersion = "dev"
	buildCommit  = "none"
	buildDate    = "unknown"
	buildOs      = "unknown"
	buildArch    = "unknown"
)

func main() {
	structName := flag.String("struct", "", "name of struct to use for metadata")
	output := flag.String("output", "metadata.yaml", "output file path")
	versionFlag := flag.Bool("V", false, "Print version information")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("omni-metagen-go version %s\n", buildVersion)
		fmt.Printf("commit: %s\n", buildCommit)
		fmt.Printf("built for %s %s at: %s\n", buildOs, buildArch, buildDate)
		os.Exit(0)
	}

	if *structName == "" {
		log.Fatal("struct name is required")
	}

	// Get the directory from environment variable set by go:generate
	dir := os.Getenv("GOFILE")
	if dir == "" {
		dir = "."
	} else {
		dir = filepath.Dir(dir)
	}

	generator, err := NewGenerator(dir)
	if err != nil {
		log.Fatal(err)
	}

	metadata, err := generator.Generate(*structName)
	if err != nil {
		log.Fatal(err)
	}

	outputDir := filepath.Dir(*output)
	if outputDir != "" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			log.Fatal(err)
		}
	}

	if err := metadata.WriteToFile(*output); err != nil {
		log.Fatal(err)
	}
}
