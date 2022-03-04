package main

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"gogql/gqbuilder"
	"gogql/models"
	"net/http"
)

func main() {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)

}

func BuildTestSchema() (graphql.Schema, error) {
	builder := gqbuilder.GetBuilder()

	ticket := builder.Object("Ticket", models.Ticket{})

	ticket.FieldFunc("tags", func(ctx context.Context, o *models.Ticket, args struct {
		Filter *models.TagFilterInput
		Order  *models.TagOrderInput
		Limit  *int
		Offset *int
	}) ([]*models.Tag, error) {
		return o.Tags, nil
	})

	mutationObj := builder.Mutation()

	mutationObj.FieldFunc("ticket_insert", func(ctx context.Context, args struct {
		Input *models.TicketInsertInput
	}) (*models.Ticket, error) {
		var tags []*models.Tag
		tags = append(tags, &models.Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &models.Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &models.Tag{Title: "Tag3", ID: "3"})

		return &models.Ticket{Title: "Ticket1", ID: "1", Tags: tags}, nil
	})

	queryObj := builder.Query()

	queryObj.FieldFunc("ticket", func(ctx context.Context, rgs struct {
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

	schema, err := builder.Build()
	return schema, err
}
