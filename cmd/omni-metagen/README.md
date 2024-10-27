# omni-gen-metadata

A metadata generator for `omni` custom command structs in Go.

## Installation

```bash
go install github.com/omnicli/sdk-go/cmd/omni-metagen@latest
```

## Usage

Add a `go:generate` directive to your code:

```go
//go:generate omni-metagen -struct=Config -o omni/my-command.metadata.yaml

// MyCommand is a command that does something
// @category my-category, my-sub-category
// @autocompletion true
type Config struct {
    // Source file to process
    Source string `omniarg:"source positional=true required=true desc=\"Source file\""`

    // Destination for output
    Output string `omniarg:"output desc=\"Output path\""`
}
```

Then run:

```bash
go generate ./...
```

## Struct Tags

The generator supports the following struct-level tags in the documentation:
- `@category`: Comma-separated list of categories
- `@autocompletion`: Set to "true" to enable autocompletion

## Field Tags

The `omniarg` tag supports the following options:

- Basic options:
  - Unnamed value: Override the parameter name
  - `type`: Parameter type (str, int, float, bool, flag, counter, enum)
  - `desc`: Parameter description
  - `required`: Set to "true" for required parameters
  - `positional`: Set to "true" for positional arguments
  - `placeholder`: Custom placeholder text

- Array and enum options:
  - `num_values`: Value range (e.g., "1..5", "1..=5", "..5")
  - `delimiter`: Value delimiter character
  - `type=enum(val1,val2)`: Enum with allowed values

- Special handling:
  - `last`: Set to "true" for final positional argument
  - `leftovers`: Set to "true" to capture remaining args
  - `allow_hyphen_values`: Allow values starting with hyphen

- Dependencies:
  - `requires`: Comma-separated list of required parameters
  - `conflicts_with`: Parameters that cannot be used together
  - `required_without`: Required if any listed param is absent
  - `required_without_all`: Required if all listed params are absent
  - `required_if_eq`: Required if param equals value
  - `required_if_eq_all`: Required if all conditions match

Use `-` as the tag value to ignore a field:
```go
internal bool `omniarg:"-"`
```
