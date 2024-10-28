package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"github.com/omnicli/sdk-go/internal/omniarg"
)

// Generator handles the metadata generation process
type Generator struct {
	dir string
}

// NewGenerator creates a new metadata generator for the given directory
func NewGenerator(dir string) *Generator {
	return &Generator{dir: dir}
}

// Generate generates metadata for the given struct
func (g *Generator) Generate(structName string) (*CommandMetadata, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, g.dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory: %w", err)
	}

	var metadata *CommandMetadata
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			if md := g.findStructMetadata(file, structName); md != nil {
				metadata = md
				break
			}
		}
	}

	if metadata == nil {
		return nil, fmt.Errorf("struct %s not found", structName)
	}

	return metadata, nil
}

// findStructMetadata locates and parses the struct metadata
func (g *Generator) findStructMetadata(file *ast.File, structName string) *CommandMetadata {
	var metadata *CommandMetadata

	ast.Inspect(file, func(n ast.Node) bool {
		// First check for GenDecl
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		// Check if it's a type declaration
		if genDecl.Tok != token.TYPE {
			return true
		}

		// Look through the specs for our struct
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if typeSpec.Name.Name != structName {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			metadata = &CommandMetadata{
				ArgParser:      true,
				Autocompletion: false,
			}

			structTags := parseStructTags(genDecl.Doc)
			g.applyStructTags(metadata, structTags)

			parameters := g.parseParameters(structType.Fields.List)
			if len(parameters) > 0 {
				metadata.Syntax = Syntax{Parameters: parameters}
			}
			return false
		}
		return true
	})

	return metadata
}

// parseParameters parses all parameters from a list of fields
func (g *Generator) parseParameters(fieldsList []*ast.Field) (parameters []Parameter) {
	parameters = make([]Parameter, 0)

outerLoop:
	for _, field := range fieldsList {
		var options map[string]interface{}
		var argNameOverride string
		alreadyTriedExtractingTags := false

		// Doing that loop allows us to:
		// - Skip if field has no name or is an embedded struct
		// - Create all fields if declared on the same line
		for _, fieldName := range field.Names {
			// If the field is unexported, skip it
			if !ast.IsExported(fieldName.Name) {
				continue
			}

			if !alreadyTriedExtractingTags && field.Tag != nil {
				alreadyTriedExtractingTags = true
				argNameOverride, options = omniarg.ExtractAndParseTag(field.Tag.Value)
				if argNameOverride == "-" {
					continue outerLoop
				}
			}

			// Prepare the parameter
			param := Parameter{
				Name: convertFieldNameToArgName(fieldName.Name),
				Type: inferType(field.Type),
			}

			// If we had a name override, apply it
			if argNameOverride != "" {
				param.Name = argNameOverride
			}

			// If any options, apply them
			if options != nil {
				g.applyOptions(&param, options)
			}

			// Make sure the parameter name is lowercase
			param.Name = omniarg.SanitizeArgName(param.Name, '-')
			if param.Name == "" {
				continue
			}

			// If not a positional, add the appropriate prefix
			if !param.Positional {
				if len(param.Name) == 1 {
					param.Name = "-" + param.Name
				} else {
					param.Name = "--" + param.Name
				}
			}

			// If we get here, add the parameter to the list
			parameters = append(parameters, param)
		}
	}

	return
}

// applyOptions applies the parsed options to a parameter
func (g *Generator) applyOptions(param *Parameter, options map[string]interface{}) {
	if desc, ok := options["desc"].(string); ok {
		param.Description = desc
	}
	if aliases, ok := options["aliases"].([]string); ok {
		param.Aliases = aliases
	}
	if positional, ok := options["positional"].(bool); ok {
		param.Positional = positional
	}
	if required, ok := options["required"].(bool); ok {
		param.Required = required
	}
	if placeholder, ok := options["placeholder"].(string); ok {
		param.Placeholder = placeholder
	}
	if typ, ok := options["type"].(string); ok {
		param.Type = typ
	}
	if values, ok := options["values"].([]string); ok {
		param.Values = values
	}
	if def, ok := options["default"]; ok {
		param.Default = def
	}
	if numVal, ok := options["num_values"].(string); ok {
		param.NumValues = numVal
	}
	if delimiter, ok := options["delimiter"].(string); ok {
		param.Delimiter = delimiter
	}
	if last, ok := options["last"].(bool); ok {
		param.Last = last
	}
	if leftovers, ok := options["leftovers"].(bool); ok {
		param.Leftovers = leftovers
	}
	if allowHyphen, ok := options["allow_hyphen_values"].(bool); ok {
		param.AllowHyphenValues = allowHyphen
	}
	if requires, ok := options["requires"].([]string); ok {
		param.Requires = requires
	}
	if conflicts, ok := options["conflicts_with"].([]string); ok {
		param.ConflictsWith = conflicts
	}
	if reqWithout, ok := options["required_without"].([]string); ok {
		param.RequiredWithout = reqWithout
	}
	if reqWithoutAll, ok := options["required_without_all"].([]string); ok {
		param.RequiredWithoutAll = reqWithoutAll
	}
	if reqIfEq, ok := options["required_if_eq"].(map[string]interface{}); ok {
		param.RequiredIfEq = reqIfEq
	}
	if reqIfEqAll, ok := options["required_if_eq_all"].(map[string]interface{}); ok {
		param.RequiredIfEqAll = reqIfEqAll
	}
}

func (g *Generator) applyStructTags(metadata *CommandMetadata, structTags map[string]interface{}) {
	if autocompletion, ok := structTags["autocompletion"].(bool); ok {
		metadata.Autocompletion = autocompletion
	}
	if category, ok := structTags["category"].([]string); ok {
		metadata.Category = category
	}
	if help, ok := structTags["help"].(string); ok {
		metadata.Help = help
	}
}
