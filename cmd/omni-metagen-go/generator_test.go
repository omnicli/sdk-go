package main_test

import (
	"os"
	"path/filepath"
	"testing"

	main "github.com/omnicli/sdk-go/cmd/omni-metagen-go"
	"github.com/stretchr/testify/assert"
)

func TestGenerator(t *testing.T) {
	// Create a temporary directory for our test files
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write test files
	writeTestFile(t, tmpDir, "basic.go", `
package testpkg

// BasicCmd demonstrates a simple command structure
//
// @autocompletion true
// @category test, example
// @help This is a test command
type BasicCmd struct {
	// Basic string flag
	Name string `+"`omniarg:\"desc=\\\"The name to use\\\" required=true\"`"+`

	// Multiple fields on one line
	X, Y int `+"`omniarg:\"type=int desc=\\\"Coordinates\\\"\"`"+`

	// Enum type
	Mode string `+"`omniarg:\"type=enum(fast,slow) default=fast\"`"+`

	// Array type
	Tags []string `+"`omniarg:\"delimiter=,\"`"+`

	// Complex requirements
	Output string `+"`omniarg:\"required_if_eq=format:json,type:full\"`"+`

	// Positional argument
	File string `+"`omniarg:\"positional=true last=true\"`"+`

	// Skipped field
	Skipped string `+"`omniarg:\"-\"`"+`

	// Unexported field
	hidden string
}`)

	tests := []struct {
		name           string
		structName     string
		expectError    bool
		expectedResult *main.CommandMetadata
	}{
		{
			name:        "basic command struct",
			structName:  "BasicCmd",
			expectError: false,
			expectedResult: &main.CommandMetadata{
				ArgParser:      true,
				Autocompletion: true,
				Category:       []string{"test", "example"},
				Help:           "This is a test command",
				Syntax: main.Syntax{
					Parameters: []main.Parameter{
						{
							Name:        "--name",
							Description: "The name to use",
							Required:    true,
							Type:        "str",
						},
						{
							Name:        "-x",
							Description: "Coordinates",
							Type:        "int",
						},
						{
							Name:        "-y",
							Description: "Coordinates",
							Type:        "int",
						},
						{
							Name:    "--mode",
							Type:    "enum",
							Values:  []string{"fast", "slow"},
							Default: "fast",
						},
						{
							Name:      "--tags",
							Type:      "array/str",
							Delimiter: ",",
						},
						{
							Name: "--output",
							Type: "str",
							RequiredIfEq: map[string]interface{}{
								"format": "json",
								"type":   "full",
							},
						},
						{
							Name:       "file",
							Type:       "str",
							Positional: true,
							Last:       true,
						},
					},
				},
			},
		},
		{
			name:        "non-existent struct",
			structName:  "NonExistentCmd",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := main.NewGenerator(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := generator.Generate(tt.structName)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestComplexStructs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-complex-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "complex.go", `
package testpkg

// ComplexCmd demonstrates advanced parameter configurations
type ComplexCmd struct {
	// Parameter with aliases
	Verbose bool `+"`omniarg:\"verbose aliases=v,vv desc=\\\"Enable verbose output\\\"\"`"+`

	// Parameter with multiple requirements
	Format string `+"`omniarg:\"format requires=output conflicts_with=raw required_without=template\"`"+`

	// Parameter with allow_hyphen_values
	Args []string `+"`omniarg:\"args type=array/string allow_hyphen_values=true leftovers=true\"`"+`

	// Parameter with num_values and allow_negative_numbers
	Range []int `+"`omniarg:\"range type=array/int num_values=2 allow_negative_numbers=true placeholders=\\\"MIN MAX\\\"\"`"+`
}`)

	tests := []struct {
		name           string
		structName     string
		expectedParams []main.Parameter
	}{
		{
			name:       "complex parameters",
			structName: "ComplexCmd",
			expectedParams: []main.Parameter{
				{
					Name:        "--verbose",
					Description: "Enable verbose output",
					Type:        "flag",
					Aliases:     []string{"v", "vv"},
				},
				{
					Name:            "--format",
					Type:            "str",
					Requires:        []string{"output"},
					ConflictsWith:   []string{"raw"},
					RequiredWithout: []string{"template"},
				},
				{
					Name:              "--args",
					Type:              "array/string",
					AllowHyphenValues: true,
					Leftovers:         true,
				},
				{
					Name:                 "--range",
					Type:                 "array/int",
					NumValues:            "2",
					Placeholders:         []string{"MIN", "MAX"},
					AllowNegativeNumbers: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := main.NewGenerator(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := generator.Generate(tt.structName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, tt.expectedParams, result.Syntax.Parameters)
		})
	}
}

func TestMultipleFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-multifile-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Write multiple test files
	writeTestFile(t, tmpDir, "cmd1.go", `
package testpkg

type Cmd1 struct {
	Name string `+"`omniarg:\"name\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd2.go", `
package testpkg

type Cmd2 struct {
	Age int `+"`omniarg:\"age\"`"+`
}`)

	tests := []struct {
		name       string
		structName string
		wantParam  string
		wantType   string
	}{
		{
			name:       "find struct in first file",
			structName: "Cmd1",
			wantParam:  "--name",
			wantType:   "str",
		},
		{
			name:       "find struct in second file",
			structName: "Cmd2",
			wantParam:  "--age",
			wantType:   "int",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator, err := main.NewGenerator(tmpDir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			result, err := generator.Generate(tt.structName)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result.Syntax.Parameters) != 1 {
				t.Fatalf("expected 1 parameter, got %d", len(result.Syntax.Parameters))
			}

			param := result.Syntax.Parameters[0]
			if param.Name != tt.wantParam {
				t.Errorf("parameter name = %v, want %v", param.Name, tt.wantParam)
			}
			if param.Type != tt.wantType {
				t.Errorf("parameter type = %v, want %v", param.Type, tt.wantType)
			}
		})
	}
}

func TestEmbeddedStruct(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    Database // Embedded struct as value
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--database-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--database-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestEmbeddedStructWithTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    Database `+"`omniarg:\"db\"`"+`
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--db-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--db-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestEmbeddedPointerStruct(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    *Database // Embedded struct as pointer
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--database-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--database-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestEmbeddedPointerStructWithTag(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    *Database `+"`omniarg:\"db\"`"+`
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--db-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--db-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestStructField(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    Primary Database `+"`omniarg:\"desc=\\\"primary database\\\"\"`"+`
    Secondary Database `+"`omniarg:\"other desc=\\\"secondary database\\\"\"`"+`
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--primary-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--primary-port",
			Description: "database port",
			Type:        "int",
		},
		{
			Name:        "--other-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--other-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestStructPointerField(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Database struct {
    Host string `+"`omniarg:\"desc=\\\"database host\\\"\"`"+`
    Port int    `+"`omniarg:\"desc=\\\"database port\\\"\"`"+`
}`)

	writeTestFile(t, tmpDir, "cmd.go", `
package testpkg

type Config struct {
    Primary *Database `+"`omniarg:\"desc=\\\"primary database\\\"\"`"+`
    Secondary *Database `+"`omniarg:\"other desc=\\\"secondary database\\\"\"`"+`
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Config")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--primary-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--primary-port",
			Description: "database port",
			Type:        "int",
		},
		{
			Name:        "--other-host",
			Description: "database host",
			Type:        "str",
		},
		{
			Name:        "--other-port",
			Description: "database port",
			Type:        "int",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestStackedStructs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "types.go", `
package testpkg

type Leaf struct {
    Value string `+"`omniarg:\"desc=\\\"final value\\\"\"`"+`
}

type NodeA struct {
    Leaf            // embedded struct
}

type NodeB struct {
    *NodeA          // embedded pointer struct
}

type NodeC struct {
    Next NodeB      // struct field
}

type Root struct {
    *NodeC          // embedded pointer to struct containing a field
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("Root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:        "--node-c-next-node-a-leaf-value",
			Description: "final value",
			Type:        "str",
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters)
}

func TestBasicGroupType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "generator-basic-group-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	writeTestFile(t, tmpDir, "basic_group.go", `
package testpkg

type BasicGroupCmd struct {
	// Simple string group without any tags
	Strings [][]string
	// Simple int group without any tags
	Ints [][]int
	// Simple float group without any tags
	Floats [][]float64
	// Simple bool group without any tags
	Bools [][]bool
}`)

	generator, err := main.NewGenerator(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	result, err := generator.Generate("BasicGroupCmd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedParams := []main.Parameter{
		{
			Name:             "--strings",
			Type:             "array/str",
			NumValues:        "1..",
			GroupOccurrences: true,
		},
		{
			Name:             "--ints",
			Type:             "array/int",
			NumValues:        "1..",
			GroupOccurrences: true,
		},
		{
			Name:             "--floats",
			Type:             "array/float",
			NumValues:        "1..",
			GroupOccurrences: true,
		},
		{
			Name:             "--bools",
			Type:             "array/bool",
			NumValues:        "1..",
			GroupOccurrences: true,
		},
	}

	assert.Equal(t, expectedParams, result.Syntax.Parameters,
		"Parameter definition for basic [][]string should match expected")
}

// Helper function to write test files
func writeTestFile(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file %s: %v", name, err)
	}
}
