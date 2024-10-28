package omniarg

import (
	"testing"
	"unicode"
)

func TestSanitizeArgName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		separator rune
		want      string
	}{
		{
			name:      "simple alphanumeric",
			input:     "hello123",
			separator: '-',
			want:      "hello123",
		},
		{
			name:      "with spaces",
			input:     "hello world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "with special characters",
			input:     "hello!@#$%^&*()world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "with multiple spaces",
			input:     "hello   world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "with leading special chars",
			input:     "!!!hello",
			separator: '-',
			want:      "hello",
		},
		{
			name:      "with trailing special chars",
			input:     "hello!!!",
			separator: '-',
			want:      "hello",
		},
		{
			name:      "empty string",
			input:     "",
			separator: '-',
			want:      "",
		},
		{
			name:      "only special characters",
			input:     "!@#$%^",
			separator: '-',
			want:      "",
		},
		{
			name:      "mixed case with numbers",
			input:     "Hello123World",
			separator: '-',
			want:      "hello123world",
		},
		{
			name:      "underscore separator",
			input:     "hello world",
			separator: '_',
			want:      "hello_world",
		},
		{
			name:      "unicode letters",
			input:     "héllo wörld",
			separator: '-',
			want:      "héllo-wörld",
		},
		{
			name:      "multiple different special chars",
			input:     "hello!world@test#now",
			separator: '-',
			want:      "hello-world-test-now",
		},
		{
			name:      "consecutive separators (_ to -)",
			input:     "hello____world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "consecutive separators (- to -)",
			input:     "hello---world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "no change when valid (-)",
			input:     "hello-world",
			separator: '-',
			want:      "hello-world",
		},
		{
			name:      "no change when valid (_)",
			input:     "hello_world",
			separator: '_',
			want:      "hello_world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeArgName(tt.input, tt.separator)
			if got != tt.want {
				t.Errorf("SanitizeArgName() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestSanitizeArgNameFuzz provides additional test coverage through fuzzing
func FuzzSanitizeArgName(f *testing.F) {
	// Add initial corpus
	seeds := []string{
		"hello",
		"hello world",
		"hello!@#world",
		"",
		"!!!",
		"123",
		"héllo",
	}

	for _, seed := range seeds {
		f.Add(seed, '-')
		f.Add(seed, '_')
	}

	f.Fuzz(func(t *testing.T, input string, separator rune) {
		result := SanitizeArgName(input, separator)

		// Verify invariants
		if len(input) > 0 && len(result) == 0 {
			// If result is empty, input must have contained only special characters
			for _, r := range input {
				if unicode.IsLetter(r) || unicode.IsNumber(r) {
					t.Errorf("Empty result with input containing alphanumeric: %q", input)
					return
				}
			}
		}

		// Check that result doesn't start or end with separator
		if len(result) > 0 {
			if rune(result[0]) == separator {
				t.Errorf("Result starts with separator: %q", result)
			}
			if rune(result[len(result)-1]) == separator {
				t.Errorf("Result ends with separator: %q", result)
			}
		}

		// Check that there are no consecutive separators
		prev := rune(0)
		for _, r := range result {
			if r == separator && prev == separator {
				t.Errorf("Found consecutive separators in result: %q", result)
			}
			prev = r
		}
	})
}
