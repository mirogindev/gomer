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
	builder := gqbuilder.GetBuilder()

	query := builder.Query()

	query.FieldResolver("topics", func(ctx context.Context) ([]*Topic, error) {
		var topics []*Topic

		topics = append(topics, &Topic{Title: "Topic1", ID: 1})
		topics = append(topics, &Topic{Title: "Topic2", ID: 2})
		topics = append(topics, &Topic{Title: "Topic3", ID: 3})

		return topics, nil
	})
}
```

This is an example which defines a schema with single `topic`  mutation field
and a `FieldResolver` method which returns the created `topic`

```go
    mutation := builder.Mutation()

    mutation.FieldResolver("topic_insert", func(ctx context.Context, 
		args struct { Input *TopicInsertInput }) (*Topic, error) {
	topic := &Topic{
		ID:    args.Input.ID,
		Title: args.Input.Title,
	}
		return topic, nil
	})
}
```

This is an example which defines a schema with single `new_topics`  subscription field
and a `FieldSubscription` method which returns the new `topics`


```go
	builder := gqbuilder.GetBuilder()

	subscription := builder.Subscription()

	subscription.FieldSubscription("new_topics",Topic{}, func(ctx context.Context, c chan interface{}) {
		var i int64

		for {
			i++

			topic := Topic{ID: i,Title: fmt.Sprintf("%d", i)}

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
}

```