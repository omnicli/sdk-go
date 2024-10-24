// Package omnicli provides the Go SDK for building Omni commands.
//
// This package offers various utilities and helpers that make it easier to work
// with Omni's features from within Go. Currently, it focuses on argument parsing,
// but future versions will include additional functionality for working with other
// Omni features.
//
// # Argument Parsing
//
// The primary feature currently available is the ability to parse Omni CLI arguments
// from environment variables into Go structs. The package supports various types
// including strings, booleans, integers, and floats, both as single values and arrays.
//
// Example usage:
//
//	type Config struct {
//	    // Fields are automatically mapped to kebab-case CLI arguments
//	    InputFile string    // maps to --input-file
//	    LogFile   *string   // maps to --log-file, optional
//	    Workers   []string  // maps to --workers (array)
//
//	    // Use tags for custom names or to skip fields
//	    DBHost    string    `omni:"db_host"`  // custom name
//	    Internal  string    `omni:"-"`        // skip this field
//	}
//
//	func main() {
//	    var cfg Config
//	    args, err := omnicli.ParseArgs(&cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	}
//
// # Field Naming
//
// By default, struct field names are converted from CamelCase to kebab-case for
// matching CLI arguments. For example:
//   - InputFile -> input_file
//   - LogFile -> log_file
//   - DBHost -> db_host
//   - OOMScore -> oom_score
//   - SuperHTTPServer -> super_http_server
//
// Custom names can be specified using the `omni` struct tag:
//
//	type Config struct {
//	    Host string `omni:"db_replica"` // maps to --db-replica
//	}
//
// Fields can be excluded from parsing using the `-` tag value:
//
//	type Config struct {
//	    Internal string `omni:"-"` // will be skipped
//	}
//
// # Optional Values
//
// Optional values should be declared as pointers. These will be nil when not set:
//
//	type Config struct {
//	    LogFile *string // nil when --log-file is not provided
//	    Workers *int    // nil when --workers is not provided
//	}
//
// # Array Values
//
// Array values are supported for all basic types:
//
//	type Config struct {
//	    Hosts []string  // --hosts value1 value2 value3
//	    Ports []int     // --ports 8080 8081 8082
//	}
//
// For the latest documentation and updates, visit:
// https://github.com/omnicli/sdk-go
package omnicli
