package reqparse_test

import (
	"strconv"
	"testing"

	"github.com/berk-karaal/reqparse"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPointer[T any](v T) *T {
	return &v
}

func TestParseQuery(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	t.Run("invalid target argument", func(t *testing.T) {
		t.Parallel()

		var s string
		err := reqparse.ParseQuery(map[string][]string{}, &s, nil)
		require.ErrorIs(t, err, reqparse.ErrInvalidQueryTarget)
	})

	t.Run("invalid target field type", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"name": {"John"},
			"age":  {"25"},
		}

		type MyStruct struct {
			Name string `query:"name"`
			Age  uint   `query:"age"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.ErrorIs(t, err, reqparse.ErrInvalidQueryFieldType)
		require.EqualError(t, err, "field type is not allowed for query parsing: Age (uint)")
	})

	t.Run("invalid slice target field type", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"param1": {"value1", "value2"},
			"param2": {"1", "2"},
		}

		type MyStruct struct {
			Param1 []string `query:"param1"`
			Param2 []uint   `query:"param2"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.ErrorIs(t, err, reqparse.ErrInvalidQueryFieldType)
		require.EqualError(t, err, "field type is not allowed for query parsing: Param2 ([]uint)")
	})

	t.Run("invalid pointer target field type", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{}

		type MyStruct struct {
			Name *string `query:"name"`
			Age  *uint   `query:"age"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.ErrorIs(t, err, reqparse.ErrInvalidQueryFieldType)
		require.EqualError(t, err, "field type is not allowed for query parsing: Age (*uint)")
	})

	t.Run("query params happy path", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"name":      {"John"},
			"age":       {"30"},
			"is_active": {"true"},
			"weight":    {"70.5"},
			"roles":     {"admin", "user"},
			"page":      {"1"},
		}

		type MyStruct struct {
			Name     string   `query:"name"`
			Age      int      `query:"age"`
			IsActive bool     `query:"is_active"`
			Weight   float64  `query:"weight"`
			Roles    []string `query:"roles"`
			Page     *int     `query:"page"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.NoError(t, err)
		assert.Equal(t, MyStruct{
			Name:     "John",
			Age:      30,
			IsActive: true,
			Weight:   70.5,
			Roles:    []string{"admin", "user"},
			Page:     newPointer(1),
		}, s)
	})

	t.Run("query tag not found", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"name": {"John"},
			"age":  {"25"},
		}

		type MyStruct struct {
			Name string `query:"name"`
			Age  int
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.ErrorIs(t, err, reqparse.ErrQueryTagNotFound)
		assert.EqualError(t, err, "query tag not found for struct field: Age")
	})

	t.Run("query params required validation error", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{}

		type MyStruct struct {
			Name     string  `query:"name"`
			Age      int     `query:"age"`
			IsActive bool    `query:"is_active"`
			Weight   float64 `query:"weight"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		var validationError *reqparse.QueryValidationError
		require.ErrorAs(t, err, &validationError)
		assert.Equal(t, reqparse.QueryValidationError{
			FieldErrors: map[string][]string{
				"name": {
					"field is required",
				},
				"age": {
					"field is required",
				},
				"is_active": {
					"field is required",
				},
				"weight": {
					"field is required",
				},
			},
			StructErrors: []string{},
		}, *validationError)
	})

	t.Run("query params type casting error", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"age":       {"do"},
			"is_active": {"not"},
			"weight":    {"panic"},
		}

		type MyStruct struct {
			Age      int     `query:"age"`
			IsActive bool    `query:"is_active"`
			Weight   float64 `query:"weight"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		var validationError *reqparse.QueryValidationError
		require.ErrorAs(t, err, &validationError)
		assert.Equal(t, reqparse.QueryValidationError{
			FieldErrors: map[string][]string{
				"age": {
					"must be a valid integer",
				},
				"is_active": {
					"must be a valid boolean",
				},
				"weight": {
					"must be a valid float",
				},
			},
			StructErrors: []string{},
		}, *validationError)
	})

	t.Run("get the first value", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"name":      {"John", "Doe"},
			"age":       {"25", "40"},
			"is_active": {"false", "true"},
			"weight":    {"70.5", "80.5"},
			"roles":     {"admin", "user"},
			"page":      {"1", "2"},
		}

		type MyStruct struct {
			Name     string   `query:"name"`
			Age      int      `query:"age"`
			IsActive bool     `query:"is_active"`
			Weight   float64  `query:"weight"`
			Roles    []string `query:"roles"`
			Page     *int     `query:"page"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.NoError(t, err)
		assert.Equal(t, MyStruct{
			Name:     "John",
			Age:      25,
			IsActive: false,
			Weight:   70.5,
			Roles:    []string{"admin", "user"},
			Page:     newPointer(1),
		}, s)
	})

	t.Run("default value", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{}

		type MyStruct struct {
			Name     string   `query:"name"      default:"John"`
			Age      int      `query:"age"       default:"25"`
			IsActive bool     `query:"is_active" default:"true"`
			Weight   float64  `query:"weight"    default:"60.5"`
			Roles    []string `query:"roles"     default:"admin,user"`
			Page     *int     `query:"page"      default:"1"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.NoError(t, err)
		assert.Equal(t, MyStruct{
			Name:     "John",
			Age:      25,
			IsActive: true,
			Weight:   60.5,
			Roles:    []string{"admin", "user"},
			Page:     newPointer(1),
		}, s)
	})

	t.Run("slice params", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"param1": {"value1"},
			"param2": {"value2", "value3", "value4"},

			"param6": {"1"},
			"param7": {"2", "3", "4"},

			"param11": {"1.1"},
			"param12": {"2.2", "3.3", "4.4"},

			"param16": {"false"},
			"param17": {"false", "true", "false"},
		}

		type MyStruct struct {
			Param1 []string `query:"param1"`
			Param2 []string `query:"param2"`
			Param3 []string `query:"param3" default:"abcd"`
			Param4 []string `query:"param4" default:"abcd,efgh,ijkl"`
			Param5 []string `query:"param5"`

			Param6  []int `query:"param6"`
			Param7  []int `query:"param7"`
			Param8  []int `query:"param8"  default:"10"`
			Param9  []int `query:"param9"  default:"10,20,30"`
			Param10 []int `query:"param10"`

			Param11 []float64 `query:"param11"`
			Param12 []float64 `query:"param12"`
			Param13 []float64 `query:"param13" default:"10.1"`
			Param14 []float64 `query:"param14" default:"10.1,20.2,30.3"`
			Param15 []float64 `query:"param15"`

			Param16 []bool `query:"param16"`
			Param17 []bool `query:"param17"`
			Param18 []bool `query:"param18" default:"true"`
			Param19 []bool `query:"param19" default:"true,false,true"`
			Param20 []bool `query:"param20"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.NoError(t, err)
		assert.Equal(t, MyStruct{
			Param1: []string{"value1"},
			Param2: []string{"value2", "value3", "value4"},
			Param3: []string{"abcd"},
			Param4: []string{"abcd", "efgh", "ijkl"},
			Param5: []string{},

			Param6:  []int{1},
			Param7:  []int{2, 3, 4},
			Param8:  []int{10},
			Param9:  []int{10, 20, 30},
			Param10: []int{},

			Param11: []float64{1.1},
			Param12: []float64{2.2, 3.3, 4.4},
			Param13: []float64{10.1},
			Param14: []float64{10.1, 20.2, 30.3},
			Param15: []float64{},

			Param16: []bool{false},
			Param17: []bool{false, true, false},
			Param18: []bool{true},
			Param19: []bool{true, false, true},
			Param20: []bool{},
		}, s)
	})

	t.Run("slice params type casting validation error", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"param1": {"value1", "value2"},
			"param2": {"hmmm", "mmmm"},
			"param4": {"hmmmmm", "aaaaaa"},
		}

		type MyStruct struct {
			Param1 []int     `query:"param1"`
			Param2 []bool    `query:"param2"`
			Param3 []string  `query:"param3"`
			Param4 []float64 `query:"param4"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		var validationError *reqparse.QueryValidationError
		require.ErrorAs(t, err, &validationError)
		assert.Equal(t, reqparse.QueryValidationError{
			FieldErrors: map[string][]string{
				"param1": {
					"(Index: 0) must be a valid integer",
					"(Index: 1) must be a valid integer",
				},
				"param2": {
					"(Index: 0) must be a valid boolean",
					"(Index: 1) must be a valid boolean",
				},
				"param4": {
					"(Index: 0) must be a valid float",
					"(Index: 1) must be a valid float",
				},
			},
			StructErrors: []string{},
		}, *validationError)
	})

	t.Run("pointer params", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"param1": {"value1"},
			"param2": {"value2", "value3", "value4"},

			"param4": {"123"},
			"param5": {"456", "789", "123"},

			"param7": {"true"},
			"param8": {"false", "true", "false"},

			"param10": {"1.1"},
			"param11": {"2.2", "3.3", "4.4"},
		}

		type MyStruct struct {
			Param1  *string  `query:"param1"`
			Param2  *string  `query:"param2"`
			Param3  *string  `query:"param3"`
			Param4  *int     `query:"param4"`
			Param5  *int     `query:"param5"`
			Param6  *int     `query:"param6"`
			Param7  *bool    `query:"param7"`
			Param8  *bool    `query:"param8"`
			Param9  *bool    `query:"param9"`
			Param10 *float64 `query:"param10"`
			Param11 *float64 `query:"param11"`
			Param12 *float64 `query:"param12"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		require.NoError(t, err)
		assert.Equal(t, MyStruct{
			Param1:  newPointer("value1"),
			Param2:  newPointer("value2"),
			Param3:  nil,
			Param4:  newPointer(123),
			Param5:  newPointer(456),
			Param6:  nil,
			Param7:  newPointer(true),
			Param8:  newPointer(false),
			Param9:  nil,
			Param10: newPointer(1.1),
			Param11: newPointer(2.2),
			Param12: nil,
		}, s)
	})

	t.Run("pointer params type casting validation error", func(t *testing.T) {
		t.Parallel()

		inputQueryParams := map[string][]string{
			"param1": {"value1", "value2"},
			"param2": {"hmmm", "mmmm"},
			"param4": {"hmmmmm", "aaaaaa"},
		}

		type MyStruct struct {
			Param1 *int     `query:"param1"`
			Param2 *bool    `query:"param2"`
			Param3 *string  `query:"param3"`
			Param4 *float64 `query:"param4"`
		}

		var s MyStruct
		err := reqparse.ParseQuery(inputQueryParams, &s, nil)

		var validationError *reqparse.QueryValidationError
		require.ErrorAs(t, err, &validationError)
		assert.Equal(t, reqparse.QueryValidationError{
			FieldErrors: map[string][]string{
				"param1": {
					"must be a valid integer",
				},
				"param2": {
					"must be a valid boolean",
				},
				"param4": {
					"must be a valid float",
				},
			},
			StructErrors: []string{},
		}, *validationError)
	})

	t.Run("QueryValidationError string representation", func(t *testing.T) {
		t.Parallel()

		validationError := reqparse.QueryValidationError{
			FieldErrors: map[string][]string{
				"param1": {
					"must be a valid integer",
				},
				"param2": {
					"must be a valid boolean",
				},
			},
			StructErrors: []string{
				"currently no struct error exists but it will be used in the future :D",
			},
		}

		// Ordering of the field errors can be different.
		ordering1 := "Parsing query parameters failed.\nStruct Errors:\n\tcurrently no struct error exists but it will be used in the future :D\nField Errors:\n\tparam1:\n\t\tmust be a valid integer\n\tparam2:\n\t\tmust be a valid boolean\n" //nolint:lll
		ordering2 := "Parsing query parameters failed.\nStruct Errors:\n\tcurrently no struct error exists but it will be used in the future :D\nField Errors:\n\tparam2:\n\t\tmust be a valid boolean\n\tparam1:\n\t\tmust be a valid integer\n" //nolint:lll

		errorText := validationError.Error()
		if errorText != ordering1 && errorText != ordering2 {
			t.Errorf(
				"string representation is not as expected options. Please see the expected options in test code. Got: %s",
				strconv.Quote(errorText),
			)
		}
	})
}
