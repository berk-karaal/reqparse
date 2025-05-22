package reqparse

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrInvalidQueryTarget = errors.New( //nolint:revive
		"target argument must be a non-nil pointer to a struct",
	)
	ErrInvalidQueryFieldType = errors.New("field type is not allowed for query parsing")
	ErrQueryTagNotFound      = errors.New("query tag not found for struct field")
)

// QueryValidationError is the error type used by [ParseQuery] function when the passed query
// parameters does not satisfy the validation rules of the struct.
type QueryValidationError struct {
	// FieldErrors contains errors for fields that have at least one error. Key is the query name of
	// the field.
	FieldErrors map[string][]string

	// StructErrors contains struct level validation errors.
	//
	// INFO: This field is not implemented at the moment, so it will always be empty. When
	// implemented, it will contain struct level validation errors.
	StructErrors []string
}

func (e *QueryValidationError) Error() string {
	var errText strings.Builder

	errText.WriteString("Parsing query parameters failed.\nStruct Errors:\n")
	for _, err := range e.StructErrors { //nolint:wsl
		errText.WriteString("\t" + err + "\n")
	}

	errText.WriteString("Field Errors:\n")
	for k, v := range e.FieldErrors { //nolint:wsl
		errText.WriteString("\t" + k + ":\n")

		for _, err := range v {
			errText.WriteString("\t\t" + err + "\n")
		}
	}

	return errText.String()
}

// ParseQueryOptions is the options type for [ParseQuery]. It will be used in the future for
// adding custom validators to [ParseQuery] and other stuff.
type ParseQueryOptions struct{}

// ParseQuery parses query parameters into given struct.
// If options are nil, default options are used.
func ParseQuery(
	queryParams map[string][]string,
	target any,
	opts *ParseQueryOptions,
) error {
	if opts == nil {
		opts = &ParseQueryOptions{} //nolint:ineffassign,wastedassign
	}

	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Pointer || v.IsNil() || v.Elem().Kind() != reflect.Struct {
		return ErrInvalidQueryTarget
	}

	validationErrors := &QueryValidationError{
		FieldErrors:  make(map[string][]string),
		StructErrors: make([]string, 0),
	}

	structElem := v.Elem()

	for i := 0; i < structElem.NumField(); i++ {
		fieldv := structElem.Field(i)
		structField := structElem.Type().Field(i)

		if !isFieldTypeAllowedForQueryParsing(fieldv.Type()) {
			return fmt.Errorf(
				"%w: %s (%s)",
				ErrInvalidQueryFieldType,
				structField.Name,
				fieldv.Type(),
			)
		}

		if err := populateStructFieldFromQuery(fieldv, structField, queryParams, validationErrors); err != nil {
			return err
		}
	}

	if len(validationErrors.StructErrors) > 0 || len(validationErrors.FieldErrors) > 0 {
		return validationErrors
	}

	return nil
}

func isFieldTypeAllowedForQueryParsing(fieldType reflect.Type) bool {
	switch fieldType.Kind() { //nolint:exhaustive
	case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
		return true
	case reflect.Slice:
		switch fieldType.Elem().Kind() { //nolint:exhaustive
		case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
			return true
		default:
			return false
		}
	case reflect.Pointer:
		switch fieldType.Elem().Kind() { //nolint:exhaustive
		case reflect.String, reflect.Int, reflect.Float64, reflect.Bool:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// populateStructFieldFromQuery finds the associated query param for the struct field and sets the
// field value accordingly. It handles default values, required fields, type casting and validation
// errors.
func populateStructFieldFromQuery( //nolint:cyclop,funlen
	fieldv reflect.Value,
	structField reflect.StructField,
	queryParams map[string][]string,
	validationErrors *QueryValidationError,
) error {
	fieldQueryKey, ok := structField.Tag.Lookup("query")
	if !ok {
		return fmt.Errorf("%w: %s", ErrQueryTagNotFound, structField.Name)
	}

	values, ok := queryParams[fieldQueryKey]
	if !ok {
		fieldDefaultValue, ok := structField.Tag.Lookup("default")
		if !ok {
			switch fieldv.Kind() { //nolint:exhaustive
			case reflect.Slice:
				// If default value is not specified for slice field which is not present in the
				// query params, set an empty slice.
				fieldv.Set(reflect.MakeSlice(fieldv.Type(), 0, 0))
			case reflect.Pointer:
				// If default value is not specified for pointer field which is not present in the
				// query params, set nil.
				fieldv.Set(reflect.Zero(fieldv.Type()))
			default:
				// If default value is not specified for other type of field which is not present in
				// the query params, add a validation error to indicate that the field is required.
				validationErrors.FieldErrors[fieldQueryKey] = append(
					validationErrors.FieldErrors[fieldQueryKey], "field is required",
				)
			}

			return nil
		}

		if fieldv.Kind() == reflect.Slice {
			values = strings.Split(fieldDefaultValue, ",")
		} else {
			values = []string{fieldDefaultValue}
		}
	}

	// Set the field value by the query values
	structFieldKind := fieldv.Kind()
	switch structFieldKind { //nolint:exhaustive
	case reflect.Slice:
		setSliceFieldValue(fieldv, values, fieldQueryKey, validationErrors)

	case reflect.Pointer:
		setPointerFieldValue(fieldv, values, fieldQueryKey, validationErrors)

	case reflect.String:
		fieldv.SetString(values[0])

	case reflect.Int:
		i, err := strconv.Atoi(values[0])
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid integer",
			)
			break
		}

		fieldv.SetInt(int64(i))

	case reflect.Float64:
		f, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid float",
			)
			break
		}

		fieldv.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid boolean",
			)
			break
		}

		fieldv.SetBool(b)
	}

	return nil
}

func setSliceFieldValue( //nolint:cyclop
	fieldv reflect.Value,
	values []string,
	fieldQueryKey string,
	validationErrors *QueryValidationError,
) {
	sliceElementKind := fieldv.Type().Elem().Kind()

	switch sliceElementKind { //nolint:exhaustive
	case reflect.String:
		fieldv.Set(reflect.ValueOf(values))

	case reflect.Int:
		newSlice := make([]int, len(values))
		for i, v := range values {
			intValue, err := strconv.Atoi(v)
			if err != nil {
				validationErrors.FieldErrors[fieldQueryKey] = append(
					validationErrors.FieldErrors[fieldQueryKey],
					"(Index: "+strconv.Itoa(i)+") must be a valid integer",
				)
			}

			newSlice[i] = intValue
		}

		fieldv.Set(reflect.ValueOf(newSlice))

	case reflect.Float64:
		newSlice := make([]float64, len(values))
		for i, v := range values {
			floatValue, err := strconv.ParseFloat(v, 64)
			if err != nil {
				validationErrors.FieldErrors[fieldQueryKey] = append(
					validationErrors.FieldErrors[fieldQueryKey],
					"(Index: "+strconv.Itoa(i)+") must be a valid float",
				)
			}

			newSlice[i] = floatValue
		}

		fieldv.Set(reflect.ValueOf(newSlice))

	case reflect.Bool:
		newSlice := make([]bool, len(values))
		for i, v := range values {
			boolValue, err := strconv.ParseBool(v)
			if err != nil {
				validationErrors.FieldErrors[fieldQueryKey] = append(
					validationErrors.FieldErrors[fieldQueryKey],
					"(Index: "+strconv.Itoa(i)+") must be a valid boolean",
				)
			}

			newSlice[i] = boolValue
		}

		fieldv.Set(reflect.ValueOf(newSlice))
	}
}

func setPointerFieldValue(
	fieldv reflect.Value,
	values []string,
	fieldQueryKey string,
	validationErrors *QueryValidationError,
) {
	pointerElementKind := fieldv.Type().Elem().Kind()

	switch pointerElementKind { //nolint:exhaustive
	case reflect.String:
		fieldv.Set(reflect.New(fieldv.Type().Elem()))
		fieldv.Elem().SetString(values[0])

	case reflect.Int:
		i, err := strconv.Atoi(values[0])
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid integer",
			)
			return
		}

		fieldv.Set(reflect.New(fieldv.Type().Elem()))
		fieldv.Elem().SetInt(int64(i))

	case reflect.Float64:
		f, err := strconv.ParseFloat(values[0], 64)
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid float",
			)
			return
		}

		fieldv.Set(reflect.New(fieldv.Type().Elem()))
		fieldv.Elem().SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(values[0])
		if err != nil {
			validationErrors.FieldErrors[fieldQueryKey] = append(
				validationErrors.FieldErrors[fieldQueryKey], "must be a valid boolean",
			)
			return
		}

		fieldv.Set(reflect.New(fieldv.Type().Elem()))
		fieldv.Elem().SetBool(b)
	}
}
