package omniarg

import (
	"strings"
	"unicode"
)

// sanitizeArgName applies the same logic as what omni will apply to convert
// the argument name. Depending on the provided separator, it will:
// - replace any non-alphanumeric characters with the separator
// - remove any leading/trailing separators
// - replace any conjoined separators with a single separator
func SanitizeArgName(name string, separator rune) string {
	var result []rune

	for i, r := range name {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			result = append(result, r)
		} else {
			if i > 0 && len(result) > 0 && result[len(result)-1] != separator {
				result = append(result, separator)
			}
		}
	}

	if len(result) > 0 && result[len(result)-1] == separator {
		result = result[:len(result)-1]
	}

	return strings.ToLower(string(result))
}
