package omnicli_test

import (
	"os"
	"testing"

	omnicli "github.com/omnicli/sdk-go"
)

func TestFieldNaming(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	// Set up environment for all naming test cases
	os.Setenv("OMNI_ARG_LIST", "log_file db_host oom_score valid_oom_value different_name")

	os.Setenv("OMNI_ARG_LOG_FILE_TYPE", "str")
	os.Setenv("OMNI_ARG_LOG_FILE_VALUE", "test.log")

	os.Setenv("OMNI_ARG_DB_HOST_TYPE", "str")
	os.Setenv("OMNI_ARG_DB_HOST_VALUE", "localhost")

	os.Setenv("OMNI_ARG_OOM_SCORE_TYPE", "int")
	os.Setenv("OMNI_ARG_OOM_SCORE_VALUE", "42")

	os.Setenv("OMNI_ARG_VALID_OOM_VALUE_TYPE", "bool")
	os.Setenv("OMNI_ARG_VALID_OOM_VALUE_VALUE", "true")

	os.Setenv("OMNI_ARG_DIFFERENT_NAME_TYPE", "str")
	os.Setenv("OMNI_ARG_DIFFERENT_NAME_VALUE", "custom")

	var cfg struct {
		LogFile       string // should map to log_file
		DBHost        string // should map to db_host
		OOMScore      int    // should map to oom_score
		ValidOOMValue bool   // should map to valid_oom_value
		CustomName    string `omniarg:"different_name"`
		SkipThis      string `omniarg:"-"`
		Internal      string `omniarg:"-"`
	}

	_, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test automatic name conversion
	if cfg.LogFile != "test.log" {
		t.Errorf("LogFile: expected 'test.log', got '%s'", cfg.LogFile)
	}

	if cfg.DBHost != "localhost" {
		t.Errorf("DBHost: expected 'localhost', got '%s'", cfg.DBHost)
	}

	if cfg.OOMScore != 42 {
		t.Errorf("OOMScore: expected 42, got %d", cfg.OOMScore)
	}

	if !cfg.ValidOOMValue {
		t.Error("ValidOOMValue: expected true")
	}

	// Test custom name via tag
	if cfg.CustomName != "custom" {
		t.Errorf("CustomName: expected 'custom', got '%s'", cfg.CustomName)
	}

	// Test skipped fields
	if cfg.SkipThis != "" {
		t.Error("SkipThis should not be set")
	}

	if cfg.Internal != "" {
		t.Error("Internal should not be set")
	}
}

func TestCaseSensitivity(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	os.Setenv("OMNI_ARG_LIST", "TestArg UPPER_ARG lower_arg")

	os.Setenv("OMNI_ARG_TESTARG_TYPE", "str")
	os.Setenv("OMNI_ARG_TESTARG_VALUE", "test")

	os.Setenv("OMNI_ARG_UPPER_ARG_TYPE", "str")
	os.Setenv("OMNI_ARG_UPPER_ARG_VALUE", "upper")

	os.Setenv("OMNI_ARG_LOWER_ARG_TYPE", "str")
	os.Setenv("OMNI_ARG_LOWER_ARG_VALUE", "lower")

	var cfg struct {
		Testarg  string
		UpperArg string
		LowerArg string
	}

	args, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test struct field values
	if cfg.Testarg != "test" {
		t.Errorf("Testarg: expected 'test', got '%s'", cfg.Testarg)
	}
	if cfg.UpperArg != "upper" {
		t.Errorf("UpperArg: expected 'upper', got '%s'", cfg.UpperArg)
	}
	if cfg.LowerArg != "lower" {
		t.Errorf("LowerArg: expected 'lower', got '%s'", cfg.LowerArg)
	}

	// Test direct Args access (should be case insensitive)
	if val, ok := args.GetString("testarg"); !ok || val != "test" {
		t.Errorf("GetString(testarg): expected 'test', got '%s'", val)
	}
	if val, ok := args.GetString("UPPER_ARG"); !ok || val != "upper" {
		t.Errorf("GetString(UPPER_ARG): expected 'upper', got '%s'", val)
	}
	if val, ok := args.GetString("lower_arg"); !ok || val != "lower" {
		t.Errorf("GetString(lower_arg): expected 'lower', got '%s'", val)
	}
}

func TestComplexNaming(t *testing.T) {
	cleanup := cleanEnv(t)
	defer cleanup()

	type ComplexConfig struct {
		XMLFile       string // should become xml_file
		JsonApiConfig string // should become json_api_config
		SSHKeyFile    string // should become ssh_key_file
		OauthToken    string // should become oauth_token
		LastID        int    // should become last_id
		MaxRPS        int    // should become max_rps
		EnableTLSv12  bool   `omniarg:"enable_tls_v12"`
		CustomName    string `omniarg:"my_custom_name"`
		SkipMe        string `omniarg:"-"`
	}

	os.Setenv("OMNI_ARG_LIST", "xml_file json_api_config ssh_key_file oauth_token last_id max_rps enable_tls_v12 my_custom_name")

	os.Setenv("OMNI_ARG_XML_FILE_TYPE", "str")
	os.Setenv("OMNI_ARG_XML_FILE_VALUE", "config.xml")

	os.Setenv("OMNI_ARG_JSON_API_CONFIG_TYPE", "str")
	os.Setenv("OMNI_ARG_JSON_API_CONFIG_VALUE", "api.json")

	os.Setenv("OMNI_ARG_SSH_KEY_FILE_TYPE", "str")
	os.Setenv("OMNI_ARG_SSH_KEY_FILE_VALUE", "id_rsa")

	os.Setenv("OMNI_ARG_OAUTH_TOKEN_TYPE", "str")
	os.Setenv("OMNI_ARG_OAUTH_TOKEN_VALUE", "token123")

	os.Setenv("OMNI_ARG_LAST_ID_TYPE", "int")
	os.Setenv("OMNI_ARG_LAST_ID_VALUE", "42")

	os.Setenv("OMNI_ARG_MAX_RPS_TYPE", "int")
	os.Setenv("OMNI_ARG_MAX_RPS_VALUE", "100")

	os.Setenv("OMNI_ARG_ENABLE_TLS_V12_TYPE", "bool")
	os.Setenv("OMNI_ARG_ENABLE_TLS_V12_VALUE", "true")

	os.Setenv("OMNI_ARG_MY_CUSTOM_NAME_TYPE", "str")
	os.Setenv("OMNI_ARG_MY_CUSTOM_NAME_VALUE", "custom")

	var cfg ComplexConfig
	_, err := omnicli.ParseArgs(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Test all field values
	tests := []struct {
		got      string
		expected string
		field    string
	}{
		{cfg.XMLFile, "config.xml", "XMLFile"},
		{cfg.JsonApiConfig, "api.json", "JsonApiConfig"},
		{cfg.SSHKeyFile, "id_rsa", "SSHKeyFile"},
		{cfg.OauthToken, "token123", "OauthToken"},
		{cfg.CustomName, "custom", "CustomName"},
	}

	for _, tt := range tests {
		if tt.got != tt.expected {
			t.Errorf("%s: expected '%s', got '%s'", tt.field, tt.expected, tt.got)
		}
	}

	if cfg.LastID != 42 {
		t.Errorf("LastID: expected 42, got %d", cfg.LastID)
	}

	if cfg.MaxRPS != 100 {
		t.Errorf("MaxRPS: expected 100, got %d", cfg.MaxRPS)
	}

	if !cfg.EnableTLSv12 {
		t.Error("EnableTLSv12: expected true")
	}

	if cfg.SkipMe != "" {
		t.Error("SkipMe should not be set")
	}
}
