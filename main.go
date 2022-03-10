package main

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"github.com/mirogindev/gomer/gqbuilder"
	"github.com/mirogindev/gomer/models"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func main() {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
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

func BuildTestSchema() (graphql.Schema, error) {
	builder := gqbuilder.GetBuilder()

	ticket := builder.Object("Ticket", models.Ticket{})

	ticket.FieldResolver("tags", func(ctx context.Context, o *models.Ticket, args struct {
		Filter *models.TagFilterInput
		Order  *models.TagOrderInput
		Limit  *int
		Offset *int
	}) ([]*models.Tag, error) {
		return o.Tags, nil
	})

	mutationObj := builder.Mutation()

	mutationObj.FieldResolver("ticket_insert", func(ctx context.Context, args struct {
		Input *models.TicketInsertInput
	}) (*models.Ticket, error) {
		var tags []*models.Tag
		tags = append(tags, &models.Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &models.Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &models.Tag{Title: "Tag3", ID: "3"})

		return &models.Ticket{Title: "Ticket1", ID: "1", Tags: tags}, nil
	})

	queryObj := builder.Query()

	queryObj.FieldResolver("ticket", func(ctx context.Context, args struct {
		Filter *models.TicketFilterInput
		Order  *models.TagOrderInput
		Limit  *int
		Offset *int
	}) ([]*models.Ticket, error) {
		var tickets []*models.Ticket
		var tags []*models.Tag
		tags = append(tags, &models.Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &models.Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &models.Tag{Title: "Tag3", ID: "3"})

		tickets = append(tickets, &models.Ticket{Title: "Ticket1", ID: "1", Tags: tags})
		tickets = append(tickets, &models.Ticket{Title: "Ticket2", ID: "2", Tags: tags})
		tickets = append(tickets, &models.Ticket{Title: "Ticket3", ID: "3", Tags: tags})
		return tickets, nil
	})

	subObj := builder.Subscription()

	subObj.FieldSubscription("test_sub", models.Ticket{}, func(ctx context.Context, c chan interface{}, args struct {
		Filter *models.TicketFilterInput
		Order  *models.TagOrderInput
		Limit  *int
		Offset *int
	}) {
		var i int

		for {
			i++

			ticket := models.Ticket{ID: fmt.Sprintf("%d", i), Number: i}

			select {
			case <-ctx.Done():
				log.Println("[RootSubscription] [Subscribe] subscription canceled")
				close(c)
				return
			default:
				c <- ticket
			}

			time.Sleep(200 * time.Millisecond)

			if i == 10 {
				close(c)
				return
			}
		}
	})

	schema, err := builder.Build()
	return schema, err
}
