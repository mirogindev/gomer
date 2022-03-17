# Gomer
Gomer is a Go framework for easy generation GraphQL API based on reflection
and graphql-go library

Supports: queries, mutations and realtime subscriptions.

### Getting Started

To install the library, run:
```bash
go get github.com/mirogindev/gomer
```

This is a single example which defines a schema with single `topics` field
and a `FieldResolver` method which returns the array of `topic`

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/mirogindev/gomer/gqbuilder"
	"log"
)

type Topic struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
}

func main() {
	builder := gqbuilder.GetBuilder()

	query := builder.Query()

	query.FieldResolver("topics", func(ctx context.Context) ([]*Topic, error) {
		var topics []*Topic

		topics = append(topics, &Topic{Title: "Topic1", ID: 1})
		topics = append(topics, &Topic{Title: "Topic2", ID: 2})
		topics = append(topics, &Topic{Title: "Topic3", ID: 3})

		return topics, nil
	})

	schema, err := builder.Build()

	if err != nil {
		panic(err)
	}

	queryStr := `
		{
			topics { title, id }
		}
	`

	params := graphql.Params{Schema: schema, RequestString: queryStr}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("Failed to execute operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON)

}
```

