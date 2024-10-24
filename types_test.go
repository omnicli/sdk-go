package omnicli_test

import (
	"errors"
	"os"
	"strings"
	"testing"

	omnicli "github.com/omnicli/sdk-go"
)

func TestBooleanValues(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	testCases := map[string]struct {
		value    string
		expected bool
	}{
		"flag1": {"true", true},
		"flag2": {"false", false},
		"flag3": {"True", true},
		"flag4": {"False", false},
		"flag5": {"tRuE", true},
		"flag6": {"fAlSe", false},
	}

	args := make([]string, 0, len(testCases))
	for name := range testCases {
		args = append(args, name)
	}

	os.Setenv("OMNI_ARG_LIST", strings.Join(args, " "))
	for name, tc := range testCases {
		os.Setenv("OMNI_ARG_"+strings.ToUpper(name)+"_TYPE", "bool")
		os.Setenv("OMNI_ARG_"+strings.ToUpper(name)+"_VALUE", tc.value)
	}

	var cfg struct {
		Flag1 bool
		Flag2 bool
		Flag3 bool
		Flag4 bool
		Flag5 bool
		Flag6 bool
	}

	parsedArgs, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test through struct
	if cfg.Flag1 != true || cfg.Flag2 != false {
		t.Error("Boolean values not parsed correctly in struct")
	}

	// Test through Args methods
	for name, tc := range testCases {
		if val, ok := parsedArgs.GetBool(name); !ok || val != tc.expected {
			t.Errorf("GetBool(%s) = %v, want %v", name, val, tc.expected)
		}
	}
}

func TestInvalidValues(t *testing.T) {
	tests := []struct {
		name      string
		argType   string
		value     string
		errorType interface{}
	}{
		{
			name:      "invalid_int",
			argType:   "int",
			value:     "not_a_number",
			errorType: &omnicli.InvalidIntegerValueError{},
		},
		{
			name:      "invalid_float",
			argType:   "float",
			value:     "not_a_number",
			errorType: &omnicli.InvalidFloatValueError{},
		},
		{
			name:      "invalid_bool",
			argType:   "bool",
			value:     "not_a_boolean",
			errorType: &omnicli.InvalidBooleanValueError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := cleanEnv(t)
			defer cleanup()

			os.Setenv("OMNI_ARG_LIST", tt.name)
			os.Setenv("OMNI_ARG_"+strings.ToUpper(tt.name)+"_TYPE", tt.argType)
			os.Setenv("OMNI_ARG_"+strings.ToUpper(tt.name)+"_VALUE", tt.value)

			_, err := omnicli.ParseArgs()
			if err == nil {
				t.Fatalf("Expected error for invalid %s value", tt.argType)
			}
			if !errors.As(err, &tt.errorType) {
				t.Errorf("Expected error type %T, got %T", tt.errorType, err)
			}
		})
	}
}

func TestOptionalValues(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "str_val int_val bool_val float_val")
	os.Setenv("OMNI_ARG_STR_VAL_TYPE", "str")
	os.Setenv("OMNI_ARG_INT_VAL_TYPE", "int")
	os.Setenv("OMNI_ARG_BOOL_VAL_TYPE", "bool")
	os.Setenv("OMNI_ARG_FLOAT_VAL_TYPE", "float")
	// Deliberately not setting any values

	var cfg struct {
		StrVal   *string
		IntVal   *int
		BoolVal  *bool
		FloatVal *float64
	}

	_, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cfg.StrVal != nil {
		t.Error("Expected nil string pointer for unset value")
	}
	if cfg.IntVal != nil {
		t.Error("Expected nil int pointer for unset value")
	}
	if cfg.BoolVal != nil {
		t.Error("Expected nil bool pointer for unset value")
	}
	if cfg.FloatVal != nil {
		t.Error("Expected nil float pointer for unset value")
	}
}

// More type-specific tests to follow...
