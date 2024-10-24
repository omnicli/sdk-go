package omnicli

import "testing"

func TestToParamName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Simple cases
		{
			"LogFile",
			"log_file",
		},
		{
			"Name",
			"name",
		},

		// Acronym cases
		{
			"UserID",
			"user_id",
		},
		{
			"UserI",
			"user_i",
		},
		{
			"IDNumber",
			"id_number",
		},

		// Special sequences
		{
			"OOMReason",
			"oom_reason",
		},
		{
			"ValidOOMReason",
			"valid_oom_reason",
		},

		// Mixed patterns
		{
			"EnableTLSV12Support",
			"enable_tlsv12_support",
		},
		{
			"userID",
			"user_id",
		},

		// Edge cases
		{
			"",
			"",
		},
		{
			"A",
			"a",
		},
		{
			"already_snake",
			"already_snake",
		},

		// Common examples
		{
			"HTTPConfig",
			"http_config",
		},
		{
			"DBConnection",
			"db_connection",
		},
		{
			"OauthToken",
			"oauth_token",
		},
		{
			"MaxRPSLimit",
			"max_rps_limit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toParamName(tt.input)
			if got != tt.expected {
				t.Errorf("toParamName(%q) = %q, want %q",
					tt.input, got, tt.expected)
			}
		})
	}
}

func TestToParamNameConsistency(t *testing.T) {
	// Test that repeated calls with the same input produce the same output
	inputs := []string{
		"UserID",
		"OOMKilled",
		"EnableTLSv12",
		"MaxRPSLimit",
	}

	for _, input := range inputs {
		first := toParamName(input)
		second := toParamName(input)
		if first != second {
			t.Errorf("Inconsistent results for %q: first=%q, second=%q",
				input, first, second)
		}

		// Verify that converting an already converted name doesn't change it
		converted := toParamName(first)
		if converted != first {
			t.Errorf("Converting already converted name %q changed it: got %q",
				first, converted)
		}
	}
}
