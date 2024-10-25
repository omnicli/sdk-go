package omnicli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// typeConverter is a generic interface for type conversion functions.
// Implementations of this interface handle the conversion of string values
// to their appropriate Go types.
type typeConverter[T any] interface {
	Convert(string) (T, error)
}

// stringConverter implements typeConverter for strings.
type stringConverter struct{}

func (c stringConverter) Convert(s string) (string, error) {
	return s, nil
}

// boolConverter implements typeConverter for booleans.
type boolConverter struct{}

func (c boolConverter) Convert(s string) (bool, error) {
	switch strings.ToLower(s) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, &InvalidBooleanValueError{fmt.Sprintf("expected 'true' or 'false', got '%s'", s)}
	}
}

// intConverter implements typeConverter for integers.
type intConverter struct{}

func (c intConverter) Convert(s string) (int, error) {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0, &InvalidIntegerValueError{fmt.Sprintf("expected integer, got '%s'", s)}
	}
	return val, nil
}

// floatConverter implements typeConverter for floats.
type floatConverter struct{}

func (c floatConverter) Convert(s string) (float64, error) {
	val, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, &InvalidFloatValueError{fmt.Sprintf("expected float, got '%s'", s)}
	}
	return val, nil
}

// Args represents the parsed arguments with type-specific storage.
// Each map stores pointers to values, where nil indicates a declared but unset value.
type Args struct {
	// Track declared arguments and their types
	declaredArgs map[string]string // maps arg name to its type

	// Store values as pointers - nil means declared but not set
	strings      map[string]*string
	bools        map[string]*bool
	ints         map[string]*int
	floats       map[string]*float64
	stringSlices map[string][]*string
	boolSlices   map[string][]*bool
	intSlices    map[string][]*int
	floatSlices  map[string][]*float64
}

// NewArgs creates a new Args instance with initialized maps.
// This is typically not called directly, as ParseArgs will create
// and return an Args instance.
func NewArgs() *Args {
	return &Args{
		declaredArgs: make(map[string]string),
		strings:      make(map[string]*string),
		bools:        make(map[string]*bool),
		ints:         make(map[string]*int),
		floats:       make(map[string]*float64),
		stringSlices: make(map[string][]*string),
		boolSlices:   make(map[string][]*bool),
		intSlices:    make(map[string][]*int),
		floatSlices:  make(map[string][]*float64),
	}
}

// GetString returns a string value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetString(name string) (string, bool) {
	ptr, ok := a.strings[strings.ToLower(name)]
	if !ok || ptr == nil {
		return "", ok
	}
	return *ptr, true
}

// GetBool returns a bool value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetBool(name string) (bool, bool) {
	ptr, ok := a.bools[strings.ToLower(name)]
	if !ok || ptr == nil {
		return false, ok
	}
	return *ptr, true
}

// GetInt returns an int value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetInt(name string) (int, bool) {
	ptr, ok := a.ints[strings.ToLower(name)]
	if !ok || ptr == nil {
		return 0, ok
	}
	return *ptr, true
}

// GetFloat returns a float value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetFloat(name string) (float64, bool) {
	ptr, ok := a.floats[strings.ToLower(name)]
	if !ok || ptr == nil {
		return 0, ok
	}
	return *ptr, true
}

// GetStringSlice returns a slice of string values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetStringSlice(name string) ([]string, bool) {
	ptr, ok := a.stringSlices[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([]string, 0), ok
	}
	result := make([]string, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = ""
		} else {
			result[i] = *p
		}
	}
	return result, true
}

// GetBoolSlice returns a slice of bool values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetBoolSlice(name string) ([]bool, bool) {
	ptr, ok := a.boolSlices[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([]bool, 0), ok
	}
	result := make([]bool, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = false
		} else {
			result[i] = *p
		}
	}
	return result, true
}

// GetIntSlice returns a slice of int values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetIntSlice(name string) ([]int, bool) {
	ptr, ok := a.intSlices[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([]int, 0), ok
	}
	result := make([]int, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = 0
		} else {
			result[i] = *p
		}
	}
	return result, true
}

// GetFloatSlice returns a slice of float values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetFloatSlice(name string) ([]float64, bool) {
	ptr, ok := a.floatSlices[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([]float64, 0), ok
	}
	result := make([]float64, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = 0
		} else {
			result[i] = *p
		}
	}
	return result, true
}

// GetAllArgs returns all declared arguments
func (a *Args) GetAllArgs() map[string]interface{} {
	result := make(map[string]interface{})
	for name, typ := range a.declaredArgs {
		switch typ {
		case "bool":
			if val, ok := a.GetBool(name); ok {
				result[name] = val
			}
		case "int":
			if val, ok := a.GetInt(name); ok {
				result[name] = val
			}
		case "float":
			if val, ok := a.GetFloat(name); ok {
				result[name] = val
			}
		default:
			if val, ok := a.GetString(name); ok {
				result[name] = val
			}
		}
	}
	return result
}

// parseTypeInfo parses the type string into base type and indicates if it's a slice.
// Returns baseType, arraySize, hasSize where hasSize indicates if a size was specified (even if it's 0).
func parseTypeInfo(typeStr string) (string, int, bool) {
	parts := strings.Split(typeStr, "/")
	if len(parts) == 2 {
		size, err := strconv.Atoi(parts[1])
		if err != nil {
			return parts[0], 0, false
		}
		return parts[0], size, true
	}
	return typeStr, 0, false
}

// getArgList gets the list of available arguments from OMNI_ARG_LIST environment variable.
func getArgList() ([]string, error) {
	argListStr, exists := os.LookupEnv("OMNI_ARG_LIST")
	if !exists {
		return nil, &ArgListMissingError{}
	}

	args := strings.Fields(argListStr)
	for i := range args {
		args[i] = strings.ToLower(args[i])
	}
	return args, nil
}

// getArgValue retrieves a single argument value from environment variables.
func getArgValue[T any](
	argName string,
	index *int,
	converter typeConverter[T],
) (*T, error) {
	var key string
	if index != nil {
		key = fmt.Sprintf("OMNI_ARG_%s_VALUE_%d", strings.ToUpper(argName), *index)
	} else {
		key = fmt.Sprintf("OMNI_ARG_%s_VALUE", strings.ToUpper(argName))
	}

	value, exists := os.LookupEnv(key)
	if !exists {
		return nil, nil
	}

	converted, err := converter.Convert(value)
	if err != nil {
		return nil, err
	}
	return &converted, nil
}

// handleValue processes a value based on type information.
func handleValue[T any](
	args *Args,
	argName string,
	isSlice bool,
	arraySize int,
	converter typeConverter[T],
	storeSingle func(*Args, string, *T),
	storeSlice func(*Args, string, []*T),
) error {
	if isSlice {
		values := make([]*T, arraySize)
		if arraySize > 0 {
			for i := 0; i < arraySize; i++ {
				idx := i
				val, err := getArgValue(argName, &idx, converter)
				if err != nil {
					return err
				}
				values[i] = val // val might be nil, which is what we want
			}
		}
		storeSlice(args, argName, values)
	} else {
		val, err := getArgValue(argName, nil, converter)
		if err != nil {
			return err
		}
		storeSingle(args, argName, val)
	}
	return nil
}

// validateFieldType checks if the struct field type matches the declared argument type.
func (a *Args) validateFieldType(field reflect.StructField, declaredType string) error {
	baseType := field.Type
	isPtr := baseType.Kind() == reflect.Ptr
	if isPtr {
		baseType = baseType.Elem()
	}

	isSlice := baseType.Kind() == reflect.Slice
	if isSlice {
		baseType = baseType.Elem()
	}

	expectedType := ""
	switch baseType.Kind() {
	case reflect.String:
		expectedType = "str"
	case reflect.Bool:
		expectedType = "bool"
	case reflect.Int:
		expectedType = "int"
	case reflect.Float64:
		expectedType = "float"
	default:
		return fmt.Errorf("unsupported field type: %v", baseType.Kind())
	}

	parts := strings.Split(declaredType, "/")
	if parts[0] != expectedType {
		return &TypeMismatchError{
			fieldName:    field.Name,
			expectedType: expectedType,
			receivedType: parts[0],
		}
	}

	hasSize := len(parts) > 1
	if hasSize != isSlice {
		if isSlice {
			return fmt.Errorf("field %q is a slice but argument is not", field.Name)
		}
		return fmt.Errorf("field %q is not a slice but argument is", field.Name)
	}

	return nil
}

// Fill populates a struct with values from the parsed arguments.
// The struct fields are matched with argument names based on their name or 'omniarg' tag.
// Field names are converted to lowercase for matching.
// It returns an error if any field cannot be filled or if types don't match.
func (a *Args) Fill(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("argument must be a non-nil pointer to a struct")
	}

	strct := val.Elem()
	if strct.Kind() != reflect.Struct {
		return fmt.Errorf("argument must be a pointer to a struct")
	}

	structType := strct.Type()

	for i := 0; i < strct.NumField(); i++ {
		field := strct.Field(i)
		fieldType := structType.Field(i)

		if !field.CanSet() {
			continue
		}

		argName := ""
		if tag, ok := fieldType.Tag.Lookup("omniarg"); ok {
			if tag == "-" {
				continue // Skip this field
			}

			argName = tag
		} else {
			argName = toParamName(fieldType.Name)
		}

		declaredType, exists := a.declaredArgs[argName]
		if !exists {
			return fmt.Errorf("error in %s: field %q: parameter %q not found",
				structType.Name(), fieldType.Name, argName)
		}

		if err := a.validateFieldType(fieldType, declaredType); err != nil {
			switch e := err.(type) {
			case *TypeMismatchError:
				return fmt.Errorf("error in %s: field %q has wrong type (expected %s, got %s)",
					structType.Name(), e.fieldName, e.expectedType, e.receivedType)
			default:
				return fmt.Errorf("error in %s: %w", structType.Name(), err)
			}
		}

		if err := a.fillField(field, argName); err != nil {
			return fmt.Errorf("error in %s: %w", structType.Name(), err)
		}
	}

	return nil
}

// FillAll attempts to fill multiple target structs.
// It stops and returns an error on the first failure.
func (a *Args) FillAll(targets ...interface{}) error {
	for _, target := range targets {
		if err := a.Fill(target); err != nil {
			return err
		}
	}
	return nil
}

// ParseArgs reads omni arguments from environment variables and optionally fills provided structs.
// If target structs are provided, it will attempt to fill each one before returning.
//
// Example:
//
//	var config Config
//	var flags Flags
//	args, err := ParseArgs(&config, &flags)
func ParseArgs(targets ...interface{}) (*Args, error) {
	argList, err := getArgList()
	if err != nil {
		return nil, err
	}

	args := NewArgs()

	for _, argName := range argList {
		// Get type from OMNI_ARG_X_TYPE env var, default to "str"
		typeStr, exists := os.LookupEnv(fmt.Sprintf("OMNI_ARG_%s_TYPE", strings.ToUpper(argName)))
		if !exists {
			typeStr = "str"
		}

		args.declaredArgs[argName] = typeStr
		baseType, arraySize, isSlice := parseTypeInfo(typeStr)

		// Default to string type for unknown types
		switch baseType {
		case "bool":
			err = handleValue[bool](args, argName, isSlice, arraySize,
				boolConverter{},
				func(a *Args, name string, val *bool) { a.bools[name] = val },
				func(a *Args, name string, val []*bool) { a.boolSlices[name] = val })

		case "int":
			err = handleValue[int](args, argName, isSlice, arraySize,
				intConverter{},
				func(a *Args, name string, val *int) { a.ints[name] = val },
				func(a *Args, name string, val []*int) { a.intSlices[name] = val })

		case "float":
			err = handleValue[float64](args, argName, isSlice, arraySize,
				floatConverter{},
				func(a *Args, name string, val *float64) { a.floats[name] = val },
				func(a *Args, name string, val []*float64) { a.floatSlices[name] = val })

		default: // Including "str" and any unknown types
			err = handleValue[string](args, argName, isSlice, arraySize,
				stringConverter{},
				func(a *Args, name string, val *string) { a.strings[name] = val },
				func(a *Args, name string, val []*string) { a.stringSlices[name] = val })
		}

		if err != nil {
			return nil, err
		}
	}

	// If targets were provided, fill them all
	if len(targets) > 0 {
		if err := args.FillAll(targets...); err != nil {
			return nil, err
		}
	}

	return args, nil
}

// fillField handles filling a single field with proper nil handling
func (a *Args) fillField(field reflect.Value, argName string) error {
	fieldType := field.Type()
	kind := fieldType.Kind()

	isTargetPtr := kind == reflect.Ptr
	if isTargetPtr {
		fieldType = fieldType.Elem()
		kind = fieldType.Kind()
	}

	if kind == reflect.Slice {
		return a.fillSliceField(field, fieldType.Elem().Kind(), argName)
	}

	return a.fillSingleField(field, kind, argName, isTargetPtr)
}

// fillSingleField handles single value fields
func (a *Args) fillSingleField(field reflect.Value, kind reflect.Kind, argName string, isTargetPtr bool) error {
	var parsedValue interface{}

	switch kind {
	case reflect.String:
		parsedValue = a.strings[argName]
	case reflect.Bool:
		parsedValue = a.bools[argName]
	case reflect.Int:
		parsedValue = a.ints[argName]
	case reflect.Float64:
		parsedValue = a.floats[argName]
	default:
		return fmt.Errorf("unsupported field type: %v", kind)
	}

	if parsedValue == nil {
		// Set to zero value for the field type, which either
		// sets the pointer to nil or the value to zero
		field.Set(reflect.Zero(field.Type()))
	} else if isTargetPtr {
		// For pointer fields, set directly to the parsed value pointer
		field.Set(reflect.ValueOf(parsedValue))
	} else {
		// Safely get the value from the pointer
		ptrVal := reflect.ValueOf(parsedValue)
		if !ptrVal.IsNil() {
			field.Set(ptrVal.Elem())
		} else {
			// If the pointer is nil, set to zero value
			field.Set(reflect.Zero(field.Type()))
		}
	}

	return nil
}

// fillSliceField handles slice fields
func (a *Args) fillSliceField(field reflect.Value, elemKind reflect.Kind, argName string) error {
	var ptrSlice interface{}

	switch elemKind {
	case reflect.String:
		ptrSlice = a.stringSlices[argName]
	case reflect.Bool:
		ptrSlice = a.boolSlices[argName]
	case reflect.Int:
		ptrSlice = a.intSlices[argName]
	case reflect.Float64:
		ptrSlice = a.floatSlices[argName]
	default:
		return fmt.Errorf("unsupported slice element type: %v", elemKind)
	}

	if ptrSlice == nil {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	// Create a new slice of the appropriate type
	sliceVal := reflect.ValueOf(ptrSlice)
	newSlice := reflect.MakeSlice(field.Type(), sliceVal.Len(), sliceVal.Len())

	// Copy values, handling nil pointers appropriately
	for i := 0; i < sliceVal.Len(); i++ {
		elemPtr := sliceVal.Index(i).Interface()
		if elemPtr == nil || reflect.ValueOf(elemPtr).IsNil() {
			// For nil pointers, set zero value
			newSlice.Index(i).Set(reflect.Zero(newSlice.Index(i).Type()))
		} else {
			ptrValue := reflect.ValueOf(elemPtr)
			if ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
				newSlice.Index(i).Set(ptrValue.Elem())
			}
		}
	}

	field.Set(newSlice)
	return nil
}
