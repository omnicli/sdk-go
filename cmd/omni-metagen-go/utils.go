package main

import (
	"fmt"
	"go/ast"
	"strings"
)

// parseStructTags parses all struct-level tags from documentation
func parseStructTags(doc *ast.CommentGroup) (options map[string]interface{}) {
	if doc == nil {
		return
	}

	options = make(map[string]interface{})
	currentOption := ""

	for _, line := range strings.Split(doc.Text(), "\n") {
		line = strings.TrimSpace(line)

		// Check if the line is a tag, and if so, update the current option
		if strings.HasPrefix(line, "@") {
			parts := strings.SplitN(line, " ", 2)
			currentOption = strings.TrimPrefix(parts[0], "@")
			if len(parts) < 2 {
				continue
			}

			line = strings.TrimSpace(parts[1])
			if line == "" {
				continue
			}
		}

		// Skip the line if we don't have a current option
		if currentOption == "" {
			continue
		}

		switch currentOption {
		case "category":
			// For the category, we split the value on commas, we trim spaces,
			// and we store the result as a slice of strings. If the value is empty,
			// we skip it. If the category is already set, we append the new values.
			newCategories := make([]string, 0)
			for _, category := range strings.Split(line, ",") {
				category = strings.TrimSpace(category)
				if category != "" {
					newCategories = append(newCategories, category)
				}
			}

			if len(newCategories) > 0 {
				if categories, ok := options["category"]; ok {
					options["category"] = append(categories.([]string), newCategories...)
				} else {
					options["category"] = newCategories
				}
			}
		case "autocompletion":
			// For autocompletion, we set the value to true if the value is "true",
			// otherwise we set it to false.
			if strings.TrimSpace(line) == "true" {
				options["autocompletion"] = true
			} else {
				options["autocompletion"] = false
			}
		case "help":
			// For the help, we append the line to the existing help text.
			if help, ok := options["help"]; ok {
				options["help"] = fmt.Sprintf("%s\n%s", help, line)
			} else {
				options["help"] = line
			}
		default:
			// For all other options, we store the value as is.
			options[currentOption] = line
		}
	}

	// If the help is set, trim spaces around it
	if help, ok := options["help"].(string); ok {
		options["help"] = strings.TrimSpace(help)
	}

	return options
}

func inferType(expr ast.Expr) (string, bool, error) {
	baseType, nestLevel, err := inferTypeWithNesting(expr, 0)
	if err != nil {
		return "", false, err
	}

	if nestLevel > 2 {
		return "", false, fmt.Errorf("too many nested arrays")
	}

	if nestLevel > 0 {
		switch baseType {
		case "flag":
			// Arrays of flags are not supported, but arrays of bools are
			baseType = "bool"
		case "counter":
			return "", false, fmt.Errorf("arrays of counters are not supported")
		}

		groupOccurrences := nestLevel > 1
		baseType = fmt.Sprintf("array/%s", baseType)
		return baseType, groupOccurrences, nil
	}

	return baseType, false, nil
}

// inferTypeWithNesting infers the parameter type from a Go AST expression
func inferTypeWithNesting(expr ast.Expr, nestLevel int) (string, int, error) {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "bool":
			return "flag", nestLevel, nil // Default bool to flag
		case "string":
			return "str", nestLevel, nil
		case "int", "int8", "int16", "int32", "int64":
			return "int", nestLevel, nil
		case "float32", "float64":
			return "float", nestLevel, nil
		default:
			return "", nestLevel, fmt.Errorf("unsupported type %s", t.Name)
		}
	case *ast.ArrayType:
		return inferTypeWithNesting(t.Elt, nestLevel+1)
	case *ast.StarExpr:
		return inferTypeWithNesting(t.X, nestLevel)
	default:
		return "", nestLevel, fmt.Errorf("unsupported type %T", t)
	}
}

// convertFieldNameToArgName converts a field name to a parameter name,
// following the same rules as struct field names in Go, i.e. camelCase
// to kebab-case
func convertFieldNameToArgName(fieldName string) string {
	var kebab strings.Builder
	for i, r := range fieldName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			kebab.WriteByte('-')
		}
		kebab.WriteRune(r)
	}
	return strings.ToLower(kebab.String())
}
