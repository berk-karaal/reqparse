# reqparse

reqparse is a Go package for parsing request query parameters into Go structs.

:information_source: Headers and request body parsing will be added in the future.

**Table of Contents**

<!-- no toc -->
- [Installation](#installation)
- [Query Parameters](#query-parameters)
- [License](#license)

## Installation

```shell
$ go get github.com/berk-karaal/reqparse
```

## Query Parameters

`ParseQuery()` function parses the query parameters into the target struct.

Example:

```go
package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/berk-karaal/reqparse"
)

type QueryParams struct {
	Search     string   `query:"q"`                // Required field, validation error if param not present
	Page       int      `query:"page" default:"1"` // Default value if param not present
	Categories []string `query:"categories[]"`     // Empty slice if param not present
	IsFree     *bool    `query:"is_free"`          // Optional field, nil if param not present
	MaxPrice   *float64 `query:"max_price"`        // Optional field, nil if param not present
}

func HandleSearchGames(w http.ResponseWriter, r *http.Request) {
	var queryParams QueryParams
	if err := reqparse.ParseQuery(r.URL.Query(), &queryParams, nil); err != nil {
		var validationError *reqparse.QueryValidationError
		if errors.As(err, &validationError) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(validationError)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(queryParams)
}
```

Example query parameters for the above example:

```
?q=racing%20game&page=2&categories[]=action&categories[]=adventure&is_free=false&max_price=19.99
```

See [docs/query_parameters.md](docs/query_parameters.md) for more details.

## License

MIT License

