package reqparse_test

import (
	"errors"
	"fmt"
	"log"

	"github.com/berk-karaal/reqparse"
)

//nolint:wsl
func ExampleParseQuery() {
	// (*http.Request).URL.Query() returns a map[string][]string
	// You will use that .URL.Query() method to get the query parameters from the request.
	// For demo purposes, we are creating a map[string][]string directly.
	inputQueryParams := map[string][]string{
		"status":       {"deployed"},
		"page":         {"2"},
		"is_active":    {"true"},
		"categories[]": {"abc", "def", "ghi"},
	}

	type QueryParams struct {
		Status     string   `query:"status"`
		Format     string   `query:"format"       default:"json"`
		Page       int      `query:"page"         default:"1"`
		IsActive   bool     `query:"is_active"`
		Categories []string `query:"categories[]"`
		Location   *string  `query:"location"`
	}

	var queryParams QueryParams
	err := reqparse.ParseQuery(inputQueryParams, &queryParams, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Status     : %#v\n", queryParams.Status)
	fmt.Printf("Format     : %#v\n", queryParams.Format)
	fmt.Printf("Page       : %#v\n", queryParams.Page)
	fmt.Printf("IsActive   : %#v\n", queryParams.IsActive)
	fmt.Printf("Categories : %#v\n", queryParams.Categories)
	fmt.Printf("Location   : %#v\n", queryParams.Location)

	// Output:
	// Status     : "deployed"
	// Format     : "json"
	// Page       : 2
	// IsActive   : true
	// Categories : []string{"abc", "def", "ghi"}
	// Location   : (*string)(nil)
}

//nolint:wsl,lll
func ExampleParseQuery_validation_error_handling() {
	// (*http.Request).URL.Query() returns a map[string][]string
	// You will use that .URL.Query() method to get the query parameters from the request.
	// For demo purposes, we are creating a map[string][]string directly.
	inputQueryParams := map[string][]string{
		"page":      {"aaa"},
		"numbers[]": {"1", "2", "aaa", "4", "a"},
	}

	type QueryParams struct {
		Status  string `query:"status"`
		Page    int    `query:"page"`
		Numbers []int  `query:"numbers[]"`
	}

	var queryParams QueryParams
	err := reqparse.ParseQuery(inputQueryParams, &queryParams, nil)
	if err != nil {
		var valErrs *reqparse.QueryValidationError
		if errors.As(err, &valErrs) {
			fmt.Printf("StructErrors             : %#v\n", valErrs.StructErrors)

			// valErrs.FieldErrors --> FieldErrors: map[string][]string{"numbers[]":[]string{"(Index: 2) must be a valid integer", "(Index: 4) must be a valid integer"}, "page":[]string{"must be a valid integer"}, "status":[]string{"field is required"}}
			// Not printing as whole map because the key order is not guaranteed.

			fmt.Printf("FieldErrors[\"page\"]      : %#v\n", valErrs.FieldErrors["page"])
			fmt.Printf("FieldErrors[\"status\"]    : %#v\n", valErrs.FieldErrors["status"])
			fmt.Printf("FieldErrors[\"numbers[]\"] : %#v\n", valErrs.FieldErrors["numbers[]"])
			return
		}

		log.Fatal(err)
	}

	// Output:
	// StructErrors             : []string{}
	// FieldErrors["page"]      : []string{"must be a valid integer"}
	// FieldErrors["status"]    : []string{"field is required"}
	// FieldErrors["numbers[]"] : []string{"(Index: 2) must be a valid integer", "(Index: 4) must be a valid integer"}
}
