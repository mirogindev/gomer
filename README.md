# Gomer
![example workflow](https://github.com/mirogindev/gomer/actions/workflows/go.yml/badge.svg)

Gomer is a Go framework for easy generation GraphQL API based on the reflection
and  the graphql-go library

Supports: queries, mutations and realtime subscriptions.

### Getting Started

To install the library, run:
```bash
go get github.com/mirogindev/gomer
```

This is an example which defines a schema with single `topics`  query field
and a `FieldResolver` method which returns the array of `topic`

```go
	builder := gqbuilder.GetBuilder()

	query := builder.Query()

	query.FieldResolver("topics", func(ctx context.Context) ([]*Topic, error) {
		var topics []*Topic

		topics = append(topics, &Topic{Title: "Topic1", ID: 1})
		topics = append(topics, &Topic{Title: "Topic2", ID: 2})
		topics = append(topics, &Topic{Title: "Topic3", ID: 3})

		return topics, nil
	})
```

This is an example which defines a schema with single `topic`  mutation field
and a `FieldResolver` method which returns the created `topic`

```go
	builder := gqbuilder.GetBuilder()
	
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
```

This is an example which defines a schema with single `new_topics`  subscription field
and a `FieldSubscription` method which returns the new `topics`


```go
	subscription := builder.Subscription()

	subscription.FieldSubscription("new_topics", Topic{}, func(ctx context.Context, c chan interface{}) {
		var i int64

		for {
			i++

			topic := Topic{ID: i, Title: fmt.Sprintf("%d", i)}

			select {
			case <-ctx.Done():
				log.Println("[RootSubscription] [Subscribe] subscription canceled")
				close(c)
				return
			default:
				c <- topic
			}

			time.Sleep(200 * time.Millisecond)

			if i == 10 {
				close(c)
				return
			}
		}

	})

```

This is the full working example

```go
package main

import (
	"context"
	"fmt"
	"github.com/graphql-go/handler"
	"github.com/mirogindev/gomer/gqbuilder"
	"log"
	"net/http"
	"time"
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

	subscription := builder.Subscription()

	subscription.FieldSubscription("new_topics", Topic{}, func(ctx context.Context, c chan interface{}) {
		var i int64

		for {
			i++

			topic := Topic{ID: i, Title: fmt.Sprintf("%d", i)}

			select {
			case <-ctx.Done():
				log.Println("[RootSubscription] [Subscribe] subscription canceled")
				close(c)
				return
			default:
				c <- topic
			}

			time.Sleep(200 * time.Millisecond)

			if i == 10 {
				close(c)
				return
			}
		}

	})

	schema, err := builder.Build()

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	sh := gqbuilder.GetSubscriptionHandler(schema)

	http.Handle("/graphql", h)

	http.HandleFunc("/subscriptions", sh.SubscriptionsHandlerFunc)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
```
