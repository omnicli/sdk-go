package omniarg

import (
	"fmt"
	"strings"
)

const omniFieldTag = "omniarg"

// ExtractAndParseTag extracts and then parses the omniarg tag into an
// arg name and options map.
//
// If the tag is empty, returns an empty string and nil
// If the tag is "-", returns "-" and nil
// If the tag is invalid, returns an empty string and nil
// If the tag is valid, returns the arg name and options map
func ExtractAndParseTag(tag string) (string, map[string]interface{}) {
	tag = strings.Trim(tag, "`")
	if tagContents, ok := extractTag(tag, omniFieldTag); ok {
		// If omniarg is empty, we're done
		if tagContents == "" {
			return "", nil
		}

		// If omniarg is "-", skip the field entirely
		if tagContents == "-" {
			return "-", nil
		}

		argNameOverride, options := ParseTag(tagContents)

		// If the argNameOverride is "-", skip the field entirely
		if argNameOverride == "-" {
			return "-", nil
		}

		return argNameOverride, options
	}

	return "", nil
}

// ParseTag parses the omniarg tag into a name and options map
func ParseTag(tag string) (string, map[string]interface{}) {
	parts := specialSplit(tag, ' ', true, true)

	options := make(map[string]interface{})
	var argName string

	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			key := strings.TrimSpace(kv[0])
			value := strings.Trim(strings.TrimSpace(kv[1]), "\"")

			switch key {
			case "aliases":
				options[key] = strings.Split(value, ",")
			case "positional", "required", "last", "leftovers", "allow_hyphen_values",
				"allow_negative_numbers", "group_occurrences":
				options[key] = value == "true"
			case "requires", "conflicts_with", "required_without", "required_without_all":
				options[key] = strings.Split(value, ",")
			case "required_if_eq", "required_if_eq_all":
				conditions := make(map[string]interface{})
				pairs := strings.Split(value, ",")
				for _, pair := range pairs {
					kv := strings.Split(pair, ":")
					if len(kv) == 2 {
						key := strings.TrimSpace(kv[0])
						value := strings.TrimSpace(kv[1])
						conditions[key] = value
					}
				}
				options[key] = conditions
			case "type":
				strType := value
				is_array := false

				if strings.HasPrefix(strType, "array/") {
					strType = strType[6:]
					is_array = true
				} else if strings.HasPrefix(strType, "[") && strings.HasSuffix(strType, "]") {
					strType = strType[1 : len(strType)-1]
					is_array = true
				}

				if strings.HasPrefix(strType, "(") && strings.HasSuffix(strType, ")") {
					options["values"] = strings.Split(strType[1:len(strType)-1], ",")
					strType = "enum"
				} else if strings.HasPrefix(strType, "enum(") && strings.HasSuffix(strType, ")") {
					options["values"] = strings.Split(strType[5:len(strType)-1], ",")
					strType = "enum"
				}

				if values, ok := options["values"].([]string); ok {
					// trim spaces for enum values
					for i, v := range values {
						values[i] = strings.TrimSpace(v)
					}
				}

				if is_array {
					options["type"] = fmt.Sprintf("array/%s", strType)
				} else {
					options["type"] = strType
				}
			case "placeholders", "placeholder":
				placeholders := strings.Split(value, " ")
				for i, placeholder := range placeholders {
					placeholders[i] = strings.TrimSpace(placeholder)
				}
				options["placeholders"] = placeholders
			default:
				options[key] = value
			}
		} else if argName == "" {
			argName = strings.TrimSpace(part)
		}
	}

	return argName, options
}

// extractTag extracts a specific tag value from a struct field tag
func extractTag(tag, key string) (string, bool) {
	for _, t := range specialSplit(tag, ' ', true, false) {
		parts := strings.SplitN(t, ":", 2)
		if parts[0] == key {
			if len(parts) == 2 {
				// Remove the " around the value _only_ if we have one on both sides
				value := parts[1]
				if strings.HasPrefix(parts[1], "\"") && strings.HasSuffix(parts[1], "\"") {
					value = value[1 : len(value)-1]
					// Unescape any escaped quotes
					value = strings.ReplaceAll(value, `\"`, `"`)
				}
				return value, true
			}
			return "", true
		}
	}
	return "", false
}

func specialSplit(s string, sep rune, respectQuotes bool, respectParens bool) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	inParens := 0
	escaped := false

	for _, r := range s {
		switch r {
		case sep:
			if inQuote || inParens > 0 {
				current.WriteRune(r)
			} else {
				parts = append(parts, current.String())
				current.Reset()
			}
		case '\\':
			escaped = !escaped
			current.WriteRune(r)
		case '"':
			if respectQuotes && !escaped {
				inQuote = !inQuote
			}
			current.WriteRune(r)
		case '(':
			if respectParens {
				inParens++
			}
			current.WriteRune(r)
		case ')':
			if respectParens {
				inParens--
			}
			current.WriteRune(r)
		default:
			current.WriteRune(r)
		}

		// Reset the escaped flag if we're not in an escape sequence
		if escaped && r != '\\' {
			escaped = false
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
