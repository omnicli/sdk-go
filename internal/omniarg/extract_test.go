package omniarg

import (
	"fmt"
	"reflect"
	"testing"
)

func TestExtractAndParseTag(t *testing.T) {
	tests := []struct {
		name          string
		tag           string
		expectedName  string
		expectedOpts  map[string]interface{}
		shouldBeEmpty bool
	}{
		{
			name:          "empty tag",
			tag:           "",
			expectedName:  "",
			expectedOpts:  nil,
			shouldBeEmpty: true,
		},
		{
			name:         "skip field with dash",
			tag:          `omniarg:"-"`,
			expectedName: "-",
			expectedOpts: nil,
		},
		{
			name:         "basic name only",
			tag:          `omniarg:"flag-name"`,
			expectedName: "flag-name",
			expectedOpts: map[string]interface{}{},
		},
		{
			name:         "name with aliases",
			tag:          `omniarg:"flag-name aliases=f,fn"`,
			expectedName: "flag-name",
			expectedOpts: map[string]interface{}{
				"aliases": []string{"f", "fn"},
			},
		},
		{
			name:         "complex tag with multiple options",
			tag:          `omniarg:"flag-name required=true type=string aliases=f,fn"`,
			expectedName: "flag-name",
			expectedOpts: map[string]interface{}{
				"required": true,
				"type":     "string",
				"aliases":  []string{"f", "fn"},
			},
		},
		{
			name:         "enum type",
			tag:          `omniarg:"color type=enum(red,green,blue)"`,
			expectedName: "color",
			expectedOpts: map[string]interface{}{
				"type":   "enum",
				"values": []string{"red", "green", "blue"},
			},
		},
		{
			name:         "array type",
			tag:          `omniarg:"numbers type=array/int"`,
			expectedName: "numbers",
			expectedOpts: map[string]interface{}{
				"type": "array/int",
			},
		},
		{
			name:         "required_if_eq condition",
			tag:          `omniarg:"output required_if_eq=format:json,type:detailed"`,
			expectedName: "output",
			expectedOpts: map[string]interface{}{
				"required_if_eq": map[string]interface{}{
					"format": "json",
					"type":   "detailed",
				},
			},
		},
		{
			name:         "multiple boolean flags",
			tag:          `omniarg:"verbose required=true positional=false last=true leftovers=false"`,
			expectedName: "verbose",
			expectedOpts: map[string]interface{}{
				"required":   true,
				"positional": false,
				"last":       true,
				"leftovers":  false,
			},
		},
		{
			name:         "requires and conflicts",
			tag:          `omniarg:"input requires=output conflicts_with=stdin"`,
			expectedName: "input",
			expectedOpts: map[string]interface{}{
				"requires":       []string{"output"},
				"conflicts_with": []string{"stdin"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, opts := ExtractAndParseTag(tt.tag)

			if name != tt.expectedName {
				t.Errorf("name = %v, want %v", name, tt.expectedName)
			}

			if tt.shouldBeEmpty {
				if opts != nil {
					t.Errorf("opts = %v, want nil", opts)
				}
				return
			}

			if !reflect.DeepEqual(opts, tt.expectedOpts) {
				t.Errorf("opts = %v, want %v", opts, tt.expectedOpts)
			}
		})
	}
}

func TestSpecialSplit(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		sep           rune
		respectQuotes bool
		respectParens bool
		expected      []string
	}{
		{
			name:          "basic split",
			input:         "a b c",
			sep:           ' ',
			respectQuotes: true,
			respectParens: true,
			expected:      []string{"a", "b", "c"},
		},
		{
			name:          "quoted strings",
			input:         `a "b c" d`,
			sep:           ' ',
			respectQuotes: true,
			respectParens: true,
			expected:      []string{"a", `"b c"`, "d"},
		},
		{
			name:          "nested parentheses",
			input:         `type=enum(a,b,c) required=true`,
			sep:           ' ',
			respectQuotes: true,
			respectParens: true,
			expected:      []string{"type=enum(a,b,c)", "required=true"},
		},
		{
			name:          "escaped quotes",
			input:         `name="first \"quote\" second" other`,
			sep:           ' ',
			respectQuotes: true,
			respectParens: true,
			expected:      []string{`name="first \"quote\" second"`, "other"},
		},
		{
			name:          "mixed quotes and parens",
			input:         `type="string" values=(a,b,"c,d")`,
			sep:           ' ',
			respectQuotes: true,
			respectParens: true,
			expected:      []string{`type="string"`, `values=(a,b,"c,d")`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := specialSplit(tt.input, tt.sep, tt.respectQuotes, tt.respectParens)
			if !reflect.DeepEqual(result, tt.expected) {
				// Pretty print the slices with indexes
				var resultStr, expectedStr string
				for i, v := range result {
					if i > 0 {
						resultStr += " "
					}
					resultStr += fmt.Sprintf("%d: %s", i, v)
				}
				for i, v := range tt.expected {
					if i > 0 {
						expectedStr += " "
					}
					expectedStr += fmt.Sprintf("%d: %s", i, v)
				}
				t.Errorf("specialSplit() =\n[%s]\nwant\n[%s]", resultStr, expectedStr)
			}
		})
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name         string
		tag          string
		expectedName string
		expectedOpts map[string]interface{}
	}{
		{
			name:         "simple tag",
			tag:          "flag-name",
			expectedName: "flag-name",
			expectedOpts: map[string]interface{}{},
		},
		{
			name:         "quoted value",
			tag:          `name help="this is help text"`,
			expectedName: "name",
			expectedOpts: map[string]interface{}{
				"help": "this is help text",
			},
		},
		{
			name:         "array notation",
			tag:          `files type=[string]`,
			expectedName: "files",
			expectedOpts: map[string]interface{}{
				"type": "array/string",
			},
		},
		{
			name:         "enum with spaces",
			tag:          `mode type=enum(fast, normal, careful)`,
			expectedName: "mode",
			expectedOpts: map[string]interface{}{
				"type":   "enum",
				"values": []string{"fast", "normal", "careful"},
			},
		},
		{
			name:         "required_without_all option",
			tag:          `output required_without_all=json-out,xml-out`,
			expectedName: "output",
			expectedOpts: map[string]interface{}{
				"required_without_all": []string{"json-out", "xml-out"},
			},
		},
		{
			name:         "allow_hyphen_values option",
			tag:          `input allow_hyphen_values=true`,
			expectedName: "input",
			expectedOpts: map[string]interface{}{
				"allow_hyphen_values": true,
			},
		},
		{
			name:         "allow_negative_numbers option",
			tag:          `number allow_negative_numbers=true`,
			expectedName: "number",
			expectedOpts: map[string]interface{}{
				"allow_negative_numbers": true,
			},
		},
		{
			name:         "group_occurrences option",
			tag:          `count group_occurrences=true`,
			expectedName: "count",
			expectedOpts: map[string]interface{}{
				"group_occurrences": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, opts := ParseTag(tt.tag)

			if name != tt.expectedName {
				t.Errorf("name = %v, want %v", name, tt.expectedName)
			}

			if !reflect.DeepEqual(opts, tt.expectedOpts) {
				t.Errorf("opts = %v, want %v", opts, tt.expectedOpts)
			}
		})
	}
}
