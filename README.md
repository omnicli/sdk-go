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
    "github.com/omnicli/sdk-go"
)

type Config struct {
    // Fields are automatically mapped to kebab-case CLI arguments
    InputFile string    // maps to --input-file
    Verbose   bool      // maps to --verbose
    LogFile   *string   // maps to --log-file, optional
    Workers   []string  // maps to --workers (array)

    // Use tags for custom names or to skip fields
    DBHost    string    `omniarg:"db_host"`  // custom name
    Internal  string    `omniarg:"-"`        // skip this field
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

## Features

- Struct-based argument parsing
- Automatic case conversion (`LogFile` will read the `log_file` argument, which comes from the `--log-file` parameter)
- Support for optional values via pointers
- Array support with type safety
- Custom argument names via tags
- Field skipping via `omniarg:"-"` tag
- Type-safe access to values
- No external dependencies

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
