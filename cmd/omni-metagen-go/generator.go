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
	dir  string
	pkgs map[string]*ast.Package // Cache packages for struct lookup
}

// NewGenerator creates a new metadata generator for the given directory
func NewGenerator(dir string) (*Generator, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parsing directory: %w", err)
	}

	return &Generator{
		dir:  dir,
		pkgs: pkgs,
	}, nil
}

// Generate generates metadata for the given struct
func (g *Generator) Generate(structName string) (*CommandMetadata, error) {
	metadata := g.findStructMetadata(structName)
	if metadata == nil {
		return nil, fmt.Errorf("struct %s not found", structName)
	}
	return metadata, nil
}

// findStructMetadata locates and parses the struct metadata
func (g *Generator) findStructMetadata(structName string) *CommandMetadata {
	st := g.findStructType(structName)
	if st == nil {
		return nil
	}

	metadata := &CommandMetadata{
		ArgParser:      true,
		Autocompletion: false,
	}

	// Parse struct level tags if available
	if doc := g.findStructDocs(structName); doc != nil {
		structTags := parseStructTags(doc)
		g.applyStructTags(metadata, structTags)
	}

	parameters := g.parseParameters(st.Fields.List, "")
	if len(parameters) > 0 {
		metadata.Syntax = Syntax{Parameters: parameters}
	}

	return metadata
}

// findStructType looks up a struct definition across all files in the packages
func (g *Generator) findStructType(typeName string) *ast.StructType {
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Files {
			var result *ast.StructType
			ast.Inspect(file, func(n ast.Node) bool {
				typeSpec, ok := n.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != typeName {
					return true
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {
					result = structType
					return false
				}
				return true
			})
			if result != nil {
				return result
			}
		}
	}
	return nil
}

// findStructDocs finds the documentation comments for a struct
func (g *Generator) findStructDocs(typeName string) *ast.CommentGroup {
	for _, pkg := range g.pkgs {
		for _, file := range pkg.Files {
			var result *ast.CommentGroup
			ast.Inspect(file, func(n ast.Node) bool {
				genDecl, ok := n.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.TYPE {
					return true
				}

				for _, spec := range genDecl.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || typeSpec.Name.Name != typeName {
						continue
					}

					result = genDecl.Doc
					return false
				}
				return true
			})
			if result != nil {
				return result
			}
		}
	}
	return nil
}

// parseParameters parses all parameters from a list of fields
func (g *Generator) parseParameters(fieldsList []*ast.Field, prefix string) (parameters []Parameter) {
	parameters = make([]Parameter, 0)

outerLoop:
	for _, field := range fieldsList {
		// Handle struct fields (both named types and inline structs)
		if nestedParams := g.handleEmbeddedStruct(field, prefix); nestedParams != nil {
			parameters = append(parameters, nestedParams...)
			continue
		}

		var options map[string]interface{}
		var argNameOverride string
		alreadyTriedExtractingTags := false

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

			// Regular field processing
			paramName := convertFieldNameToArgName(fieldName.Name)

			// If we had a name override, apply it with prefix if needed
			if argNameOverride != "" {
				paramName = argNameOverride
			}

			// Make sure the parameter name is lowercase
			paramName = omniarg.SanitizeArgName(paramName, '-')
			if paramName == "" {
				continue
			}

			// Add the prefix
			paramName = prefix + paramName

			// Handle struct fields (both named types and inline structs)
			if nestedParams := g.handleStructField(field, paramName); nestedParams != nil {
				parameters = append(parameters, nestedParams...)
				continue
			}

			param := Parameter{
				Name: paramName,
				Type: inferType(field.Type),
			}

			// If any options, apply them
			if options != nil {
				g.applyOptions(&param, options)
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

func (g *Generator) handleEmbeddedStruct(field *ast.Field, prefix string) []Parameter {
	// Embedded structs are unnamed, so we return early if there are names
	if len(field.Names) > 0 {
		return nil
	}

	structName := ""
	if field.Tag != nil {
		structName, _ = omniarg.ExtractAndParseTag(field.Tag.Value)
	}
	if structName == "-" {
		return nil
	}

	if structName == "" {
		switch t := field.Type.(type) {
		case *ast.Ident:
			// Named type from same package
			structName = convertFieldNameToArgName(t.Name)

		case *ast.SelectorExpr:
			// Imported type
			structName = convertFieldNameToArgName(t.Sel.Name)

		case *ast.StarExpr:
			// Pointer, so unwrap it and call recursively
			unwrapped := &ast.Field{
				Names: field.Names,
				Type:  t.X,
				Tag:   field.Tag,
			}
			return g.handleEmbeddedStruct(unwrapped, prefix)

		default:
			return nil
		}
	}

	structName = omniarg.SanitizeArgName(structName, '-')
	if structName == "" {
		return nil
	}
	if prefix != "" {
		structName = prefix + structName
	}

	return g.handleStructField(field, structName)
}

// handleStructField processes named struct fields (both named types and inline structs)
func (g *Generator) handleStructField(field *ast.Field, paramName string) []Parameter {
	// Get the struct fields based on the type
	var structFields []*ast.Field

	switch t := field.Type.(type) {
	case *ast.StructType:
		// Inline struct definition
		structFields = t.Fields.List

	case *ast.StarExpr:
		// Pointer, so unwrap it and call recursively
		unwrapped := &ast.Field{
			Names: field.Names,
			Type:  t.X,
			Tag:   field.Tag,
		}
		return g.handleStructField(unwrapped, paramName)

	case *ast.Ident:
		// Named type from same package
		if st := g.findStructType(t.Name); st != nil {
			structFields = st.Fields.List
		}

	case *ast.SelectorExpr:
		// Imported type
		if st := g.findStructType(t.Sel.Name); st != nil {
			structFields = st.Fields.List
		}

	default:
		// Not a struct type
		return nil
	}

	if structFields == nil {
		return nil
	}

	// Parse the struct's fields with the new prefix
	return g.parseParameters(structFields, paramName+"-")
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
	if defMissing, ok := options["default_missing_value"]; ok {
		param.DefaultMissingValue = defMissing
	}
	if numVal, ok := options["num_values"].(string); ok {
		param.NumValues = numVal
	}
	if groupOcc, ok := options["group_occurrences"].(bool); ok {
		param.GroupOccurrences = groupOcc
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
	if allowNeg, ok := options["allow_negative_numbers"].(bool); ok {
		param.AllowNegativeNumbers = allowNeg
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
