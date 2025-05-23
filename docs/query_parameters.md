# Parsing Query Parameters

- [Parsing Query Parameters](#parsing-query-parameters)
  - [ParseQuery()](#parsequery)
    - [Target Struct](#target-struct)
      - [Default Values](#default-values)
      - [Optional Fields](#optional-fields)
      - [Required Fields](#required-fields)
    - [Handling Validation Errors](#handling-validation-errors)

reqparse offers default values, required fields, optional (nil) fields and type casting for query
parameters.

## ParseQuery()

`reqparse.ParseQuery(queryParams map[string][]string, target any, opts *ParseQueryOptions) error`
function is used to parse query parameters into the target struct.

- `queryParams` argument is the input query parameters.
  - Use `(http.Request).URL.Query()` if you are using `net/http` package.
  - Use `(gin.Context).Request.URL.Query()` if you are using Gin.
  - Use `(echo.Context).Request().URL.Query()` if you are using Echo.
- `target` argument is the target struct to parse query parameters into. Make sure to pass a non-nil
pointer to a struct. See [Target Struct](#target-struct)
- `opts` argument is the options for the function. Currently unused, but will be used in the future
for adding custom validators. You can pass `nil` to use default options.

### Target Struct

Example:

```go
type QueryParams struct {
	Search     string   `query:"q"`                // Required field, validation error if param not present
	Page       int      `query:"page" default:"1"` // Default value if param not present
	Categories []string `query:"categories[]"`     // List of string values. Empty slice if param not present
	IsFree     *bool    `query:"is_free"`          // Optional field, nil if param not present
	MaxPrice   *float64 `query:"max_price"`        // Optional field, nil if param not present
}
```

Currently only `string`, `int`, `bool`, `float64`, `[]string`, `[]int`, `[]bool`, `[]float64`,
`*string`, `*int`, `*bool`, `*float64` field types are supported. Other field types will cause
`reqparse.ErrInvalidQueryFieldType` error.

Query parameter name is specified by the `query` tag. Every field must have a `query` tag. Absence
of `query` tag will cause `reqparse.ErrQueryTagNotFound` error.

#### Default Values

Default values are specified by the `default` tag. Default values are used when the query parameter
is not present.

Use comma separated values to assign default value to a slice field.

Examples:

```go
type QueryParams struct {
	Name     string   `query:"name" default:"John Doe"`
	Page     int      `query:"page" default:"1"`
	IsActive bool     `query:"is_active" default:"true"`
	MaxPrice float64  `query:"max_price" default:"100.0"`
	Roles    []string `query:"roles[]" default:"admin,user"`
}
```

#### Optional Fields

Pointer fields are optional. If a pointer field is not present in the query parameters, it will be
set to `nil`.

Also slice fields are optional. If a slice field is not present in the query parameters, it will be
set to an empty slice.

#### Required Fields

Non-pointer, non-slice fields with no default value are required. If a required field is not present
in the query parameters a validation error will be returned.

### Handling Validation Errors

```go
var queryParams QueryParams
if err := reqparse.ParseQuery(r.URL.Query(), &queryParams, nil); err != nil {
    var validationError *reqparse.QueryValidationError
    if errors.As(err, &validationError) {
        // Handle validation error, typically respond 400 or 422 status code with the validation
        // error details.
        return
    }
    // Other errors such as `reqparse.ErrInvalidQueryFieldType` or `reqparse.ErrQueryTagNotFound`
    // indicates that your code is incorrect. Typically respond 500 status code and fix the error.
    return
}
```


