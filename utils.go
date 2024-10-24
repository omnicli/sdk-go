package omnicli

import (
	"unicode"
)

// toParamName converts a struct field name to a parameter name.
// Examples:
// - LogFile -> log_file
// - OOMReason -> oom_reason
// - ValidOOMReason -> valid_oom_reason
// - ID -> id
// - UserID -> user_id
func toParamName(name string) string {
	var result []rune

	for i, r := range name {
		isUpper := unicode.IsUpper(r)

		if isUpper {
			// We need to add an underscore if:
			// - not the first character AND
			//   - (the prev character is lowercase) OR
			//   - (not the last character AND the next character is lowercase)
			if i > 0 && (unicode.IsLower(rune(name[i-1])) || (i+1 < len(name) && unicode.IsLower(rune(name[i+1])))) {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}

	return string(result)
}
