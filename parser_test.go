package omnicli_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	omnicli "github.com/omnicli/sdk-go"
)

func cleanEnv(t *testing.T) func() {
	// Store original environment
	oldEnv := make(map[string]string)
	for _, env := range os.Environ() {
		if len(env) > 9 && env[:9] == "OMNI_ARG_" {
			key := env[:strings.IndexByte(env, '=')]
			oldEnv[key] = os.Getenv(key)
		}
	}

	// Clear all OMNI_ARG variables
	for key := range oldEnv {
		os.Unsetenv(key)
	}

	// Return cleanup function
	return func() {
		// Clear any new OMNI_ARG variables
		for _, env := range os.Environ() {
			if len(env) > 9 && env[:9] == "OMNI_ARG_" {
				key := env[:strings.IndexByte(env, '=')]
				os.Unsetenv(key)
			}
		}
		// Restore original environment
		for key, value := range oldEnv {
			os.Setenv(key, value)
		}
	}
}

func TestMissingArgList(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	_, err := omnicli.ParseArgs()
	if err == nil {
		t.Fatal("Expected error for missing OMNI_ARG_LIST")
	}
	if _, ok := err.(*omnicli.ArgListMissingError); !ok {
		t.Fatalf("Expected ArgListMissingError, got %T", err)
	}
}

func TestEmptyArgList(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "")
	args, err := omnicli.ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(args.GetAllArgs()) != 0 {
		t.Error("Expected empty args for empty OMNI_ARG_LIST")
	}
}

func TestStringArgDefaults(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "test1 test2")
	os.Setenv("OMNI_ARG_TEST1_TYPE", "str")
	os.Setenv("OMNI_ARG_TEST1_VALUE", "value")
	os.Setenv("OMNI_ARG_TEST2_TYPE", "str")
	// Deliberately not setting TEST2_VALUE

	var cfg struct {
		Test1 string
		Test2 string
	}

	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test through struct
	if cfg.Test1 != "value" {
		t.Errorf("Expected Test1 = 'value', got '%s'", cfg.Test1)
	}
	if cfg.Test2 != "" {
		t.Errorf("Expected Test2 = '', got '%s'", cfg.Test2)
	}

	// Test through Args methods
	if val, ok := args.GetString("test1"); !ok || val != "value" {
		t.Errorf("Expected GetString(test1) = 'value', got '%s'", val)
	}
	if val, ok := args.GetString("test2"); !ok || val != "" {
		t.Errorf("Expected GetString(test2) = '', got '%s'", val)
	}
}

func TestArrayHandling(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "numbers strings bools floats")

	// Integer array
	os.Setenv("OMNI_ARG_NUMBERS_TYPE", "int/3")
	os.Setenv("OMNI_ARG_NUMBERS_VALUE_0", "1")
	// Deliberately skipping VALUE_1
	os.Setenv("OMNI_ARG_NUMBERS_VALUE_2", "3")

	// String array
	os.Setenv("OMNI_ARG_STRINGS_TYPE", "str/3")
	os.Setenv("OMNI_ARG_STRINGS_VALUE_0", "hello")
	// Deliberately skipping VALUE_1
	os.Setenv("OMNI_ARG_STRINGS_VALUE_2", "world")

	// Boolean array
	os.Setenv("OMNI_ARG_BOOLS_TYPE", "bool/3")
	os.Setenv("OMNI_ARG_BOOLS_VALUE_0", "true")
	// Deliberately skipping VALUE_1
	os.Setenv("OMNI_ARG_BOOLS_VALUE_2", "false")

	// Float array
	os.Setenv("OMNI_ARG_FLOATS_TYPE", "float/3")
	os.Setenv("OMNI_ARG_FLOATS_VALUE_0", "1.1")
	// Deliberately skipping VALUE_1
	os.Setenv("OMNI_ARG_FLOATS_VALUE_2", "3.3")

	var cfg struct {
		Numbers []int
		Strings []string
		Bools   []bool
		Floats  []float64
	}

	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test integers
	expectedNums := []int{1, 0, 3}
	if !reflect.DeepEqual(cfg.Numbers, expectedNums) {
		t.Errorf("Numbers = %v, want %v", cfg.Numbers, expectedNums)
	}

	// Test strings
	expectedStrs := []string{"hello", "", "world"}
	if !reflect.DeepEqual(cfg.Strings, expectedStrs) {
		t.Errorf("Strings = %v, want %v", cfg.Strings, expectedStrs)
	}

	// Test booleans
	expectedBools := []bool{true, false, false}
	if !reflect.DeepEqual(cfg.Bools, expectedBools) {
		t.Errorf("Bools = %v, want %v", cfg.Bools, expectedBools)
	}

	// Test floats
	expectedFloats := []float64{1.1, 0.0, 3.3}
	if !reflect.DeepEqual(cfg.Floats, expectedFloats) {
		t.Errorf("Floats = %v, want %v", cfg.Floats, expectedFloats)
	}

	// Test through Args methods
	if nums, ok := args.GetIntSlice("numbers"); !ok || !reflect.DeepEqual(nums, expectedNums) {
		t.Errorf("GetIntSlice(numbers) = %v, want %v", nums, expectedNums)
	}
}

func TestEmbeddedStructValue(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "user_name inner_value config_override_value config_setting")
	os.Setenv("OMNI_ARG_USER_NAME_TYPE", "str")
	os.Setenv("OMNI_ARG_USER_NAME_VALUE", "john")
	os.Setenv("OMNI_ARG_INNER_VALUE_TYPE", "int")
	os.Setenv("OMNI_ARG_INNER_VALUE_VALUE", "42")
	os.Setenv("OMNI_ARG_CONFIG_SETTING_TYPE", "bool")
	os.Setenv("OMNI_ARG_CONFIG_SETTING_VALUE", "true")
	os.Setenv("OMNI_ARG_CONFIG_OVERRIDE_VALUE_TYPE", "int")
	os.Setenv("OMNI_ARG_CONFIG_OVERRIDE_VALUE_VALUE", "765")

	type Inner struct {
		Value int
	}

	type Config struct {
		Setting bool
		Inner   Inner `omniarg:"override"`
	}

	type Outer struct {
		UserName string
		Inner    Inner
		Config   Config
	}

	var cfg Outer
	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test the filled values
	if cfg.UserName != "john" {
		t.Errorf("Expected UserName = 'john', got '%s'", cfg.UserName)
	}
	if cfg.Inner.Value != 42 {
		t.Errorf("Expected Inner.Value = 42, got %d", cfg.Inner.Value)
	}
	if !cfg.Config.Setting {
		t.Errorf("Expected Config.Setting = true, got false")
	}
	if cfg.Config.Inner.Value != 765 {
		t.Errorf("Expected Config.Inner.Value = 765, got %d", cfg.Config.Inner.Value)
	}

	// Test through Args methods
	if val, ok := args.GetString("user_name"); !ok || val != "john" {
		t.Errorf("Expected GetString(user_name) = 'john', got '%s'", val)
	}
	if val, ok := args.GetInt("inner_value"); !ok || val != 42 {
		t.Errorf("Expected GetInt(outer_inner_value) = 42, got %d", val)
	}
	if val, ok := args.GetBool("config_setting"); !ok || !val {
		t.Errorf("Expected GetBool(outer_config_setting) = true, got false")
	}
	if val, ok := args.GetInt("config_override_value"); !ok || val != 765 {
		t.Errorf("Expected GetInt(outer_config_override_value) = 765, got %d", val)
	}
}

func TestEmbeddedStructPointer(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "name settings_enabled settings_value settings_items")
	os.Setenv("OMNI_ARG_NAME_TYPE", "str")
	os.Setenv("OMNI_ARG_NAME_VALUE", "test")
	os.Setenv("OMNI_ARG_SETTINGS_ENABLED_TYPE", "bool")
	os.Setenv("OMNI_ARG_SETTINGS_ENABLED_VALUE", "true")
	os.Setenv("OMNI_ARG_SETTINGS_VALUE_TYPE", "float")
	os.Setenv("OMNI_ARG_SETTINGS_VALUE_VALUE", "3.14")
	os.Setenv("OMNI_ARG_SETTINGS_ITEMS_TYPE", "str/2")
	os.Setenv("OMNI_ARG_SETTINGS_ITEMS_VALUE_0", "item1")
	os.Setenv("OMNI_ARG_SETTINGS_ITEMS_VALUE_1", "item2")

	type Settings struct {
		Enabled bool
		Value   float64
		Items   []string
	}

	type Config struct {
		Name     string
		Settings *Settings `omniarg:"settings"`
	}

	var cfg Config
	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test the filled values
	if cfg.Name != "test" {
		t.Errorf("Expected Name = 'test', got '%s'", cfg.Name)
	}
	if cfg.Settings == nil {
		t.Fatal("Expected Settings to be non-nil")
	}
	if !cfg.Settings.Enabled {
		t.Errorf("Expected Settings.Enabled = true, got false")
	}
	if cfg.Settings.Value != 3.14 {
		t.Errorf("Expected Settings.Value = 3.14, got %f", cfg.Settings.Value)
	}
	expectedItems := []string{"item1", "item2"}
	if !reflect.DeepEqual(cfg.Settings.Items, expectedItems) {
		t.Errorf("Expected Settings.Items = %v, got %v", expectedItems, cfg.Settings.Items)
	}

	// Test through Args methods
	if val, ok := args.GetString("name"); !ok || val != "test" {
		t.Errorf("Expected GetString(name) = 'test', got '%s'", val)
	}
	if val, ok := args.GetBool("settings_enabled"); !ok || !val {
		t.Errorf("Expected GetBool(settings_enabled) = true, got false")
	}
	if val, ok := args.GetFloat("settings_value"); !ok || val != 3.14 {
		t.Errorf("Expected GetFloat(settings_value) = 3.14, got %f", val)
	}
	if items, ok := args.GetStringSlice("settings_items"); !ok || !reflect.DeepEqual(items, expectedItems) {
		t.Errorf("Expected GetStringSlice(settings_items) = %v, got %v", expectedItems, items)
	}
}

func TestGroupValues(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	// Set up group values environment
	os.Setenv("OMNI_ARG_LIST", "string_groups int_groups float_groups bool_groups")

	// String groups
	os.Setenv("OMNI_ARG_STRING_GROUPS_TYPE", "str/3/2")
	os.Setenv("OMNI_ARG_STRING_GROUPS_TYPE_0", "str/2")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_0_0", "a1")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_0_1", "a2")
	os.Setenv("OMNI_ARG_STRING_GROUPS_TYPE_1", "str/3")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_1_0", "b1")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_1_1", "b2")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_1_2", "b3")
	os.Setenv("OMNI_ARG_STRING_GROUPS_TYPE_2", "str/1")
	os.Setenv("OMNI_ARG_STRING_GROUPS_VALUE_2_0", "c1")

	// Int groups
	os.Setenv("OMNI_ARG_INT_GROUPS_TYPE", "int/2/2")
	os.Setenv("OMNI_ARG_INT_GROUPS_TYPE_0", "int/2")
	os.Setenv("OMNI_ARG_INT_GROUPS_VALUE_0_0", "1")
	os.Setenv("OMNI_ARG_INT_GROUPS_VALUE_0_1", "2")
	os.Setenv("OMNI_ARG_INT_GROUPS_TYPE_1", "int/2")
	os.Setenv("OMNI_ARG_INT_GROUPS_VALUE_1_0", "3")
	os.Setenv("OMNI_ARG_INT_GROUPS_VALUE_1_1", "4")

	// Float groups
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_TYPE", "float/2/2")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_TYPE_0", "float/2")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_VALUE_0_0", "1.1")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_VALUE_0_1", "2.2")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_TYPE_1", "float/2")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_VALUE_1_0", "3.3")
	os.Setenv("OMNI_ARG_FLOAT_GROUPS_VALUE_1_1", "4.4")

	// Bool groups
	os.Setenv("OMNI_ARG_BOOL_GROUPS_TYPE", "bool/2/2")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_TYPE_0", "bool/2")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_VALUE_0_0", "true")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_VALUE_0_1", "false")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_TYPE_1", "bool/2")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_VALUE_1_0", "false")
	os.Setenv("OMNI_ARG_BOOL_GROUPS_VALUE_1_1", "true")

	type Config struct {
		StringGroups [][]string
		IntGroups    [][]int
		FloatGroups  [][]float64
		BoolGroups   [][]bool
	}

	var cfg Config
	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test string groups
	expectedStringGroups := [][]string{
		{"a1", "a2"},
		{"b1", "b2", "b3"},
		{"c1"},
	}
	if !reflect.DeepEqual(cfg.StringGroups, expectedStringGroups) {
		t.Errorf("StringGroups = %v, want %v", cfg.StringGroups, expectedStringGroups)
	}

	// Test int groups
	expectedIntGroups := [][]int{
		{1, 2},
		{3, 4},
	}
	if !reflect.DeepEqual(cfg.IntGroups, expectedIntGroups) {
		t.Errorf("IntGroups = %v, want %v", cfg.IntGroups, expectedIntGroups)
	}

	// Test float groups
	expectedFloatGroups := [][]float64{
		{1.1, 2.2},
		{3.3, 4.4},
	}
	if !reflect.DeepEqual(cfg.FloatGroups, expectedFloatGroups) {
		t.Errorf("FloatGroups = %v, want %v", cfg.FloatGroups, expectedFloatGroups)
	}

	// Test bool groups
	expectedBoolGroups := [][]bool{
		{true, false},
		{false, true},
	}
	if !reflect.DeepEqual(cfg.BoolGroups, expectedBoolGroups) {
		t.Errorf("BoolGroups = %v, want %v", cfg.BoolGroups, expectedBoolGroups)
	}

	// Test getting groups through Args methods
	if groups, ok := args.GetStringGroups("string_groups"); !ok || !reflect.DeepEqual(groups, expectedStringGroups) {
		t.Errorf("GetStringGroups() = %v, want %v", groups, expectedStringGroups)
	}
	if groups, ok := args.GetIntGroups("int_groups"); !ok || !reflect.DeepEqual(groups, expectedIntGroups) {
		t.Errorf("GetIntGroups() = %v, want %v", groups, expectedIntGroups)
	}
	if groups, ok := args.GetFloatGroups("float_groups"); !ok || !reflect.DeepEqual(groups, expectedFloatGroups) {
		t.Errorf("GetFloatGroups() = %v, want %v", groups, expectedFloatGroups)
	}
	if groups, ok := args.GetBoolGroups("bool_groups"); !ok || !reflect.DeepEqual(groups, expectedBoolGroups) {
		t.Errorf("GetBoolGroups() = %v, want %v", groups, expectedBoolGroups)
	}
}

func TestErrorCases(t *testing.T) {
	tests := []struct {
		name      string
		setupEnv  func()
		config    interface{}
		expectErr string
	}{
		{
			name: "Invalid type format",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str/abc") // Invalid size
			},
			config:    &struct{ Test string }{},
			expectErr: "invalid type string",
		},
		{
			name: "Invalid group format",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str/2/abc") // Invalid group size
			},
			config:    &struct{ Test [][]string }{},
			expectErr: "invalid type string",
		},
		{
			name: "Type mismatch with group",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str/0/0")
			},
			config:    &struct{ Test []string }{}, // Should be [][]string
			expectErr: "is not for grouped occurrences but argument is",
		},
		{
			name: "Invalid boolean value",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "bool")
				os.Setenv("OMNI_ARG_TEST_VALUE", "invalid")
			},
			config:    &struct{ Test bool }{},
			expectErr: "expected 'true' or 'false', got 'invalid'",
		},
		{
			name: "Invalid integer value",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "int")
				os.Setenv("OMNI_ARG_TEST_VALUE", "12.34")
			},
			config:    &struct{ Test int }{},
			expectErr: "expected integer, got '12.34'",
		},
		{
			name: "Invalid float value",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "float")
				os.Setenv("OMNI_ARG_TEST_VALUE", "not-a-float")
			},
			config:    &struct{ Test float64 }{},
			expectErr: "expected float, got 'not-a-float'",
		},
		{
			name: "Missing type for group index",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str/2/2")
				// Not setting OMNI_ARG_TEST_TYPE_0
			},
			config:    &struct{ Test [][]string }{},
			expectErr: "OMNI_ARG_TEST_TYPE_0 environment variable is not set",
		},
		{
			name: "Unsupported field type",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str")
			},
			config:    &struct{ Test complex128 }{},
			expectErr: "unsupported field type",
		},
		{
			name: "Too many type segments",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str/1/2/3/4")
			},
			config:    &struct{ Test string }{},
			expectErr: "invalid type string",
		},
		{
			name: "Non-struct target",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str")
			},
			config:    new(string),
			expectErr: "must be a pointer to a struct",
		},
		{
			name: "Nil pointer target",
			setupEnv: func() {
				os.Setenv("OMNI_ARG_LIST", "test")
				os.Setenv("OMNI_ARG_TEST_TYPE", "str")
			},
			config:    (*struct{})(nil),
			expectErr: "must be a non-nil pointer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := cleanEnv(t)
			defer cleanup()

			tt.setupEnv()

			_, err := omnicli.ParseArgs(tt.config)
			if err == nil {
				t.Fatal("Expected error but got nil")
			}
			if !strings.Contains(err.Error(), tt.expectErr) {
				t.Errorf("Expected error containing %q, got %q", tt.expectErr, err.Error())
			}
		})
	}
}

func TestArgumentPointers(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "str_val str_slice str_group")

	// Single value
	os.Setenv("OMNI_ARG_STR_VAL_TYPE", "str")
	os.Setenv("OMNI_ARG_STR_VAL_VALUE", "test")

	// Slice
	os.Setenv("OMNI_ARG_STR_SLICE_TYPE", "str/2")
	os.Setenv("OMNI_ARG_STR_SLICE_VALUE_0", "a")
	os.Setenv("OMNI_ARG_STR_SLICE_VALUE_1", "b")

	// Group
	os.Setenv("OMNI_ARG_STR_GROUP_TYPE", "str/2/2")
	os.Setenv("OMNI_ARG_STR_GROUP_TYPE_0", "str/2")
	os.Setenv("OMNI_ARG_STR_GROUP_VALUE_0_0", "x")
	os.Setenv("OMNI_ARG_STR_GROUP_VALUE_0_1", "y")
	os.Setenv("OMNI_ARG_STR_GROUP_TYPE_1", "str/2")
	os.Setenv("OMNI_ARG_STR_GROUP_VALUE_1_0", "m")
	os.Setenv("OMNI_ARG_STR_GROUP_VALUE_1_1", "n")

	type Config struct {
		StrVal   *string
		StrSlice []*string
		StrGroup [][]*string
	}

	var cfg Config
	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test single value pointer
	if cfg.StrVal == nil {
		t.Error("Expected StrVal to be non-nil")
	} else if *cfg.StrVal != "test" {
		t.Errorf("Expected StrVal = 'test', got '%s'", *cfg.StrVal)
	}

	// Test slice of pointers
	if len(cfg.StrSlice) != 2 {
		t.Errorf("Expected StrSlice length 2, got %d", len(cfg.StrSlice))
	} else {
		if cfg.StrSlice[0] == nil {
			t.Error("Expected StrSlice[0] to be non-nil")
		} else if *cfg.StrSlice[0] != "a" {
			t.Errorf("Expected StrSlice[0] = 'a', got '%s'", *cfg.StrSlice[0])
		}
		if cfg.StrSlice[1] == nil {
			t.Error("Expected StrSlice[1] to be non-nil")
		} else if *cfg.StrSlice[1] != "b" {
			t.Errorf("Expected StrSlice[1] = 'b', got '%s'", *cfg.StrSlice[1])
		}
	}

	// // Test group of pointers
	// if len(cfg.StrGroup) != 2 {
	// t.Errorf("Expected StrGroup length 2, got %d", len(cfg.StrGroup))
	// } else {
	// // Test first group
	// if len(cfg.StrGroup[0]) != 2 {
	// t.Errorf("Expected StrGroup[0] length 2, got %d", len(cfg.StrGroup[0]))
	// } else {
	// if cfg.StrGroup[0][0] == nil {
	// t.Error("Expected StrGroup[0][0] to be non-nil")
	// } else if *cfg.StrGroup[0][0] != "x" {
	// t.Errorf("Expected StrGroup[0][0] = 'x', got '%s'", *cfg.StrGroup[0][0])
	// }
	// if cfg.StrGroup[0][1] == nil {
	// t.Error("Expected StrGroup[0][1] to be non-nil")
	// } else if *cfg.StrGroup[0][1] != "y" {
	// t.Errorf("Expected StrGroup[0][1] = 'y', got '%s'", *cfg.StrGroup[0][1])
	// }
	// }

	// // Test second group
	// if len(cfg.StrGroup[1]) != 2 {
	// t.Errorf("Expected StrGroup[1] length 2, got %d", len(cfg.StrGroup[1]))
	// } else {
	// if cfg.StrGroup[1][0] == nil {
	// t.Error("Expected StrGroup[1][0] to be non-nil")
	// } else if *cfg.StrGroup[1][0] != "m" {
	// t.Errorf("Expected StrGroup[1][0] = 'm', got '%s'", *cfg.StrGroup[1][0])
	// }
	// if cfg.StrGroup[1][1] == nil {
	// t.Error("Expected StrGroup[1][1] to be non-nil")
	// } else if *cfg.StrGroup[1][1] != "n" {
	// t.Errorf("Expected StrGroup[1][1] = 'n', got '%s'", *cfg.StrGroup[1][1])
	// }
	// }
	// }

	// Test also through the Args methods
	if val, ok := args.GetString("str_val"); !ok || val != "test" {
		t.Errorf("Expected GetString(str_val) = 'test', got '%s'", val)
	}

	if slice, ok := args.GetStringSlice("str_slice"); !ok || !reflect.DeepEqual(slice, []string{"a", "b"}) {
		t.Errorf("Expected GetStringSlice(str_slice) = ['a', 'b'], got %v", slice)
	}

	// if groups, ok := args.GetStringGroups("str_group"); !ok ||
	// !reflect.DeepEqual(groups, [][]string{{"x", "y"}, {"m", "n"}}) {
	// t.Errorf("Expected GetStringGroups(str_group) = [['x', 'y'], ['m', 'n']], got %v", groups)
	// }
}
