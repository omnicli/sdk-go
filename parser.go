package omnicli

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/omnicli/sdk-go/internal/omniarg"
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

// typeInfo stores information about the type of an argument.
type typeInfo struct {
	rawType   string
	baseType  string
	sliceSize int
	isSlice   bool
	isGroup   bool
}

// Args represents the parsed arguments with type-specific storage.
// Each map stores pointers to values, where nil indicates a declared but unset value.
type Args struct {
	// Track declared arguments and their types
	declaredArgs map[string]*typeInfo

	// Store values as pointers - nil means declared but not set
	// Single values
	strings map[string]*string
	bools   map[string]*bool
	ints    map[string]*int
	floats  map[string]*float64
	// Slice values
	stringSlices map[string][]*string
	boolSlices   map[string][]*bool
	intSlices    map[string][]*int
	floatSlices  map[string][]*float64
	// Grouped values, so each entry is a slice of slices
	stringGroups map[string][][]*string
	boolGroups   map[string][][]*bool
	intGroups    map[string][][]*int
	floatGroups  map[string][][]*float64
}

// NewArgs creates a new Args instance with initialized maps.
// This is typically not called directly, as ParseArgs will create
// and return an Args instance.
func NewArgs() *Args {
	return &Args{
		declaredArgs: make(map[string]*typeInfo),
		strings:      make(map[string]*string),
		bools:        make(map[string]*bool),
		ints:         make(map[string]*int),
		floats:       make(map[string]*float64),
		stringSlices: make(map[string][]*string),
		boolSlices:   make(map[string][]*bool),
		intSlices:    make(map[string][]*int),
		floatSlices:  make(map[string][]*float64),
		stringGroups: make(map[string][][]*string),
		boolGroups:   make(map[string][][]*bool),
		intGroups:    make(map[string][][]*int),
		floatGroups:  make(map[string][][]*float64),
	}
}

// Generic function to get a single value
func getSingle[T any](name string, values map[string]*T, defaultValue T) (T, bool) {
	ptr, ok := values[strings.ToLower(name)]
	if !ok || ptr == nil {
		return defaultValue, ok
	}
	return *ptr, true
}

// Generic function to get a slice
func getSlice[T any](name string, slices map[string][]*T, defaultValue T) ([]T, bool) {
	ptr, ok := slices[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([]T, 0), ok
	}
	result := make([]T, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = defaultValue
		} else {
			result[i] = *p
		}
	}
	return result, true
}

// Generic function to get groups
func getGroups[T any](name string, groups map[string][][]*T, defaultValue T) ([][]T, bool) {
	ptr, ok := groups[strings.ToLower(name)]
	if !ok {
		return nil, ok
	} else if ptr == nil {
		return make([][]T, 0), ok
	}
	result := make([][]T, len(ptr))
	for i, p := range ptr {
		if p == nil {
			result[i] = make([]T, 0)
		} else {
			result[i] = make([]T, len(p))
			for j, q := range p {
				if q == nil {
					result[i][j] = defaultValue
				} else {
					result[i][j] = *q
				}
			}
		}
	}
	return result, true
}

// GetString returns a string value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetString(name string) (string, bool) {
	return getSingle(name, a.strings, "")
}

// GetBool returns a bool value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetBool(name string) (bool, bool) {
	return getSingle(name, a.bools, false)
}

// GetInt returns an int value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetInt(name string) (int, bool) {
	return getSingle(name, a.ints, 0)
}

// GetFloat returns a float value and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetFloat(name string) (float64, bool) {
	return getSingle(name, a.floats, 0)
}

// GetStringSlice returns a slice of string values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetStringSlice(name string) ([]string, bool) {
	return getSlice(name, a.stringSlices, "")
}

// GetBoolSlice returns a slice of bool values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetBoolSlice(name string) ([]bool, bool) {
	return getSlice(name, a.boolSlices, false)
}

// GetIntSlice returns a slice of int values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetIntSlice(name string) ([]int, bool) {
	return getSlice(name, a.intSlices, 0)
}

// GetFloatSlice returns a slice of float values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetFloatSlice(name string) ([]float64, bool) {
	return getSlice(name, a.floatSlices, 0)
}

// GetStringGroups returns a slice of slices of string values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetStringGroups(name string) ([][]string, bool) {
	return getGroups(name, a.stringGroups, "")
}

// GetBoolGroups returns a slice of slices of bool values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetBoolGroups(name string) ([][]bool, bool) {
	return getGroups(name, a.boolGroups, false)
}

// GetIntGroups returns a slice of slices of int values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetIntGroups(name string) ([][]int, bool) {
	return getGroups(name, a.intGroups, 0)
}

// GetFloatGroups returns a slice of slices of float values and whether it exists.
// The boolean return value indicates whether the argument exists and is set.
func (a *Args) GetFloatGroups(name string) ([][]float64, bool) {
	return getGroups(name, a.floatGroups, 0)
}

// GetAllArgs returns all declared arguments
func (a *Args) GetAllArgs() map[string]interface{} {
	result := make(map[string]interface{})
	for name, typeInfo := range a.declaredArgs {
		switch typeInfo.baseType {
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
func parseTypeInfo(typeStr string) (*typeInfo, error) {
	parts := strings.Split(typeStr, "/")
	if len(parts) > 3 {
		return nil, &InvalidTypeStringError{typeStr}
	}

	baseType := parts[0]
	isSlice := len(parts) > 1
	isGroup := len(parts) > 2

	var sliceSize int
	if isSlice {
		convertedSliceSize, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, &InvalidTypeStringError{typeStr}
		}
		sliceSize = convertedSliceSize
	}

	if isGroup {
		if _, err := strconv.Atoi(parts[2]); err != nil {
			return nil, &InvalidTypeStringError{typeStr}
		}
	}

	return &typeInfo{
		rawType:   typeStr,
		baseType:  baseType,
		sliceSize: sliceSize,
		isSlice:   isSlice,
		isGroup:   isGroup,
	}, nil
}

// getArgType returns the declared type of an argument.
func getArgType(name string, index *int) (*typeInfo, error) {
	keyParts := []string{"OMNI_ARG", strings.ToUpper(name), "TYPE"}
	if index != nil {
		keyParts = append(keyParts, fmt.Sprintf("%d", *index))
	}
	key := strings.Join(keyParts, "_")

	typeStr, exists := os.LookupEnv(key)
	if !exists {
		return nil, &ArgTypeMissingError{name, index}
	}

	typeInfo, err := parseTypeInfo(typeStr)
	if err != nil {
		return nil, err
	}

	return typeInfo, nil
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
	group_index *int,
	converter typeConverter[T],
) (*T, error) {
	keyParts := []string{"OMNI_ARG", strings.ToUpper(argName), "VALUE"}
	if index != nil {
		keyParts = append(keyParts, fmt.Sprintf("%d", *index))
	}
	if group_index != nil {
		keyParts = append(keyParts, fmt.Sprintf("%d", *group_index))
	}
	key := strings.Join(keyParts, "_")

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
	typeInfo *typeInfo,
	converter typeConverter[T],
	storeSingle func(*Args, string, *T),
	storeSlice func(*Args, string, []*T),
	storeGroup func(*Args, string, [][]*T),
) error {
	if typeInfo.isGroup {
		return handleGroupValue(args, argName, typeInfo.sliceSize, converter, storeGroup)
	}

	if typeInfo.isSlice {
		return handleSliceValue(args, argName, typeInfo.sliceSize, converter, storeSlice)
	}

	return handleSingleValue(args, argName, converter, storeSingle)
}

func handleSingleValue[T any](
	args *Args,
	argName string,
	converter typeConverter[T],
	storeSingle func(*Args, string, *T),
) error {
	val, err := getArgValue(argName, nil, nil, converter)
	if err != nil {
		return err
	}
	storeSingle(args, argName, val)
	return nil
}

func handleSliceValue[T any](
	args *Args,
	argName string,
	sliceSize int,
	converter typeConverter[T],
	storeSlice func(*Args, string, []*T),
) error {
	values := make([]*T, sliceSize)
	for i := 0; i < sliceSize; i++ {
		idx := i

		val, err := getArgValue(argName, &idx, nil, converter)
		if err != nil {
			return err
		}

		values[i] = val // val might be nil, which is what we want
	}
	storeSlice(args, argName, values)
	return nil
}

func handleGroupValue[T any](
	args *Args,
	argName string,
	sliceSize int,
	converter typeConverter[T],
	storeGroup func(*Args, string, [][]*T),
) error {
	values := make([][]*T, sliceSize)
	for i := 0; i < sliceSize; i++ {
		idx := i

		groupTypeInfo, err := getArgType(argName, &idx)
		if err != nil {
			return err
		}
		groupSize := groupTypeInfo.sliceSize

		groupValues := make([]*T, groupSize)
		for j := 0; j < groupSize; j++ {
			groupIdx := j

			val, err := getArgValue(argName, &idx, &groupIdx, converter)
			if err != nil {
				return err
			}

			groupValues[groupIdx] = val // val might be nil, which is what we want
		}
		values[i] = groupValues
	}
	storeGroup(args, argName, values)
	return nil
}

// validateFieldType checks if the struct field type matches the declared argument type.
func (a *Args) validateFieldType(field reflect.StructField, typeInfo *typeInfo) error {
	baseType := field.Type

	isSlice := baseType.Kind() == reflect.Slice
	isGroup := false
	if isSlice {
		baseType = baseType.Elem()
		if baseType.Kind() == reflect.Slice {
			isGroup = true
			baseType = baseType.Elem()
		}
	}

	isPtr := baseType.Kind() == reflect.Ptr
	if isPtr {
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
		return fmt.Errorf("unsupported field type for %s: %v", field.Name, baseType.Kind())
	}

	if typeInfo.baseType != expectedType {
		return &TypeMismatchError{
			fieldName:    field.Name,
			expectedType: expectedType,
			receivedType: typeInfo.baseType,
		}
	}

	if typeInfo.isGroup != isGroup {
		if isGroup {
			return fmt.Errorf("field %q is for grouped occurrences but argument is not", field.Name)
		}
		return fmt.Errorf("field %q is not for grouped occurrences but argument is", field.Name)
	}

	if typeInfo.isSlice != isSlice {
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
func (a *Args) Fill(v interface{}, prefix ...string) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("argument must be a non-nil pointer to a struct")
	}

	strct := val.Elem()
	if strct.Kind() != reflect.Struct {
		return fmt.Errorf("argument must be a pointer to a struct")
	}

	structType := strct.Type()
	currentPrefix := ""
	if len(prefix) > 0 {
		currentPrefix = prefix[0]
	}

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

			argNameOverride, _ := omniarg.ParseTag(tag)
			if argNameOverride != "" {
				argName = tag
			}
		} else {
			argName = toParamName(fieldType.Name)
		}

		argName = omniarg.SanitizeArgName(argName, '_')
		if argName == "" {
			return fmt.Errorf("error in %s: field %q: missing argument name",
				structType.Name(), fieldType.Name)
		}
		argName = currentPrefix + argName

		// Handle embedded struct
		isStruct := field.Kind() == reflect.Struct
		isPtrToStruct := field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct
		if isStruct || isPtrToStruct {
			var fieldInterface interface{}
			if isStruct {
				fieldInterface = field.Addr().Interface()
			} else {
				if field.IsNil() {
					// Initialize the pointer with a new struct
					field.Set(reflect.New(field.Type().Elem()))
				}

				fieldInterface = field.Interface()
			}

			// Recursively fill embedded struct with new prefix
			if err := a.Fill(fieldInterface, argName+"_"); err != nil {
				return fmt.Errorf("error in embedded struct %s: %w", fieldType.Name, err)
			}
			continue
		}

		typeInfo, exists := a.declaredArgs[argName]
		if !exists {
			return fmt.Errorf("error in %s: field %q: parameter %q not found",
				structType.Name(), fieldType.Name, argName)
		}

		if err := a.validateFieldType(fieldType, typeInfo); err != nil {
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
		typeInfo, err := getArgType(argName, nil)
		if err != nil {
			return nil, err
		}

		args.declaredArgs[argName] = typeInfo

		// Default to string type for unknown types
		switch typeInfo.baseType {
		case "bool":
			err = handleValue[bool](args, argName, typeInfo,
				boolConverter{},
				func(a *Args, name string, val *bool) { a.bools[name] = val },
				func(a *Args, name string, val []*bool) { a.boolSlices[name] = val },
				func(a *Args, name string, val [][]*bool) { a.boolGroups[name] = val })

		case "int":
			err = handleValue[int](args, argName, typeInfo,
				intConverter{},
				func(a *Args, name string, val *int) { a.ints[name] = val },
				func(a *Args, name string, val []*int) { a.intSlices[name] = val },
				func(a *Args, name string, val [][]*int) { a.intGroups[name] = val })

		case "float":
			err = handleValue[float64](args, argName, typeInfo,
				floatConverter{},
				func(a *Args, name string, val *float64) { a.floats[name] = val },
				func(a *Args, name string, val []*float64) { a.floatSlices[name] = val },
				func(a *Args, name string, val [][]*float64) { a.floatGroups[name] = val })

		default: // Including "str" and any unknown types
			err = handleValue[string](args, argName, typeInfo,
				stringConverter{},
				func(a *Args, name string, val *string) { a.strings[name] = val },
				func(a *Args, name string, val []*string) { a.stringSlices[name] = val },
				func(a *Args, name string, val [][]*string) { a.stringGroups[name] = val })
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

	isSlice := fieldType.Kind() == reflect.Slice
	isGroup := false
	if isSlice {
		fieldType = fieldType.Elem()
		if fieldType.Kind() == reflect.Slice {
			isGroup = true
			fieldType = fieldType.Elem()
		}
	}

	isPtr := fieldType.Kind() == reflect.Ptr
	if isPtr {
		fieldType = fieldType.Elem()
	}

	kind := fieldType.Kind()

	if isGroup {
		return a.fillGroupField(field, kind, argName, isPtr)
	}

	if isSlice {
		return a.fillSliceField(field, kind, argName, isPtr)
	}

	return a.fillSingleField(field, kind, argName, isPtr)
}

// fillSingleField handles single value fields
func (a *Args) fillSingleField(field reflect.Value, elemKind reflect.Kind, argName string, isTargetPtr bool) error {
	var parsedValue interface{}

	switch elemKind {
	case reflect.String:
		parsedValue = a.strings[argName]
	case reflect.Bool:
		parsedValue = a.bools[argName]
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsedValue = a.ints[argName]
	case reflect.Float32, reflect.Float64:
		parsedValue = a.floats[argName]
	default:
		return fmt.Errorf("unsupported field type for %s: %v", argName, elemKind)
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
func (a *Args) fillSliceField(field reflect.Value, elemKind reflect.Kind, argName string, isTargetPtr bool) error {
	var ptrSlice interface{}

	switch elemKind {
	case reflect.String:
		ptrSlice = a.stringSlices[argName]
	case reflect.Bool:
		ptrSlice = a.boolSlices[argName]
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ptrSlice = a.intSlices[argName]
	case reflect.Float32, reflect.Float64:
		ptrSlice = a.floatSlices[argName]
	default:
		return fmt.Errorf("unsupported field type for %s: %v", argName, elemKind)
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
		} else if isTargetPtr {
			// For pointer fields, set directly to the parsed value pointer
			newSlice.Index(i).Set(reflect.ValueOf(elemPtr))
		} else {
			ptrValue := reflect.ValueOf(elemPtr)
			if ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
				newSlice.Index(i).Set(ptrValue.Elem())
			} else {
				// If the pointer is nil, set to zero value
				newSlice.Index(i).Set(reflect.Zero(newSlice.Index(i).Type()))
			}
		}
	}

	field.Set(newSlice)
	return nil
}

// fillGroupField handles grouped slice fields
func (a *Args) fillGroupField(field reflect.Value, elemKind reflect.Kind, argName string, isTargetPtr bool) error {
	var groupSlice interface{}

	switch elemKind {
	case reflect.String:
		groupSlice = a.stringGroups[argName]
	case reflect.Bool:
		groupSlice = a.boolGroups[argName]
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		groupSlice = a.intGroups[argName]
	case reflect.Float32, reflect.Float64:
		groupSlice = a.floatGroups[argName]
	default:
		return fmt.Errorf("unsupported field type for %s: %v", argName, elemKind)
	}

	if groupSlice == nil {
		field.Set(reflect.MakeSlice(field.Type(), 0, 0))
		return nil
	}

	// Create a new slice of slices of the appropriate type
	groupVal := reflect.ValueOf(groupSlice)
	newGroup := reflect.MakeSlice(field.Type(), groupVal.Len(), groupVal.Len())

	// Copy values for each sub-slice
	for i := 0; i < groupVal.Len(); i++ {
		subSlice := groupVal.Index(i)
		if subSlice.IsNil() {
			newGroup.Index(i).Set(reflect.MakeSlice(field.Type().Elem(), 0, 0))
			continue
		}

		newSubSlice := reflect.MakeSlice(field.Type().Elem(), subSlice.Len(), subSlice.Len())

		for j := 0; j < subSlice.Len(); j++ {
			elemPtr := subSlice.Index(j).Interface()
			if elemPtr == nil || reflect.ValueOf(elemPtr).IsNil() {
				// For nil pointers, set zero value
				newSubSlice.Index(j).Set(reflect.Zero(newSubSlice.Index(j).Type()))
			} else if isTargetPtr {
				// For pointer fields, set directly to the parsed value pointer
				newSubSlice.Index(j).Set(reflect.ValueOf(elemPtr))
			} else {
				ptrValue := reflect.ValueOf(elemPtr)
				if ptrValue.Kind() == reflect.Ptr && !ptrValue.IsNil() {
					newSubSlice.Index(j).Set(ptrValue.Elem())
				} else {
					// If the pointer is nil, set to zero value
					newSubSlice.Index(j).Set(reflect.Zero(newSubSlice.Index(j).Type()))
				}
			}
		}

		newGroup.Index(i).Set(newSubSlice)
	}

	field.Set(newGroup)
	return nil
}
