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
