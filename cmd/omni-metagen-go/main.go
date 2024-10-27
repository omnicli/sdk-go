package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
)

func main() {
	structName := flag.String("struct", "", "name of struct to use for metadata")
	output := flag.String("output", "metadata.yaml", "output file path")
	flag.Parse()

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

	generator := NewGenerator(dir)
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
