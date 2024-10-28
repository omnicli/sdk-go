# omnicli (sdk-go)

Go SDK for building Omni commands.

## Overview

`omnicli` is a Go package that provides functionality to help build commands that will be executed by Omni. It offers various utilities and helpers that make it easier to work with Omni's features from within Go.

## Installation

```bash
go get github.com/omnicli/sdk-go
```

## Features

### Argument Parsing

The SDK can read omni-parsed arguments from environment variables into Go structs:

```go
package main

import (
	"log"

	omnicli "github.com/omnicli/sdk-go"
)

//go:generate omni-metagen-go -struct=Config -output=dist/my-command.metadata.yaml

type Config struct {
	// Fields are automatically mapped to kebab-case CLI arguments
	InputFile string   // maps to --input-file
	Verbose   bool     // maps to --verbose
	LogFile   *string  // maps to --log-file, optional
	Workers   []string // maps to --workers (array)

	// Use tags for custom names or to skip fields
	DBHost   string `omniarg:"db_host"` // custom name
	Internal string `omniarg:"-"`       // skip this field
}

func main() {
	var cfg Config
	_, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		log.Fatalf("Failed to parse args: %v", err)
	}

	if cfg.Verbose {
		log.Println("Verbose mode enabled")
	}
	if cfg.InputFile != "" {
		log.Printf("Processing file: %s", cfg.InputFile)
	}
}
```

The resulting arguments can be accessed either through the populated struct or through the returned `Args` object, which provides type-safe getters for all values.

### Integration with omni

The argument parser of omni needs to be enabled for your command. This can be done as part of the [metadata](https://omnicli.dev/reference/custom-commands/path/metadata-headers) of your command, which can either be provided as a separate file:

```
your-repo
└── commands
    ├── your-command.go
    ├── your-command.sh
    └── your-command.metadata.yaml
```

```yaml
# your-command.metadata.yaml
argparser: true
```

Or as part of your command file wrapper header:

```go
#!/usr/bin/env bash
#
# argparser: true
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
go run "${DIR}"/your-command.go "$@"
```

### Metadata Generation

The SDK provides a code generation tool `omni-metagen-go` that can be used to generate metadata files for your commands. This tool will generate the metadata in YAML format based on the struct you provide. You will still need to indicate the name of the struct and the location of the output file (which should be in the same directory as your command, and named `<command>.metadata.yaml`).

You can install the tool by running:

```bash
go install github.com/omnicli/sdk-go/cmd/omni-metagen-go@latest
```

Or by using GitHub releases in omni:
```yaml
up:
  - github-releases:
      omnicli/sdk-go: latest
```

The example above shows how to setup the metadata generation in your Go code. You can then call `go generate ./...` to generate the metadata file.

## Development

To set up for development:

```bash
# Clone the repository
omni clone https://github.com/omnicli/sdk-go.git
# Install dependencies
omni up
# Run tests
omni test
```

## Requirements

- Go 1.18 or higher (for generics support)
- No additional dependencies required
