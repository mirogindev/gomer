# Gomer
Gomer is a Go framework for easy generation GraphQL API based on reflection
and graphql-go library

Supports: queries, mutations and realtime subscriptions.

### Getting Started

To install the library, run:
```bash
go get github.com/mirogindev/gomer
```

This is an example which defines a schema with single `topics`  query field
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

This is an example which defines a schema with single `topic`  mutation field
and a `FieldResolver` method which returns the created `topic`

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

type TopicInsertInput struct {
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

	mutation := builder.Mutation()

	mutation.FieldResolver("topic_insert", func(ctx context.Context, args struct {
		Input *TopicInsertInput
	}) (*Topic, error) {
		topic := &Topic{
			ID:    args.Input.ID,
			Title: args.Input.Title,
		}

		return topic, nil
	})

	schema, err := builder.Build()

	if err != nil {
		panic(err)
	}

	queryStr := `
		mutation {
			topic_insert (input: { id: 1, title: "created topic" } )  { title, id }
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