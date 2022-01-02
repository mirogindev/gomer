package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"gogql/gqbuilder"
	"gogql/models"
	"log"
	"testing"
)

func TestGetAllBudgetsWithPermissions(t *testing.T) {

	builder := gqbuilder.SchemaBuilder{}

	queryObj := builder.Query()

	queryObj.FieldFunc("ticket", func(ctx context.Context, args struct {
		Limit  *int64
		Offset *int64
	}) ([]*models.Ticket, error) {
		var tickets []*models.Ticket
		tickets = append(tickets, &models.Ticket{Title: "Ticket1", ID: "1"})
		tickets = append(tickets, &models.Ticket{Title: "Ticket2", ID: "2"})
		tickets = append(tickets, &models.Ticket{Title: "Ticket3", ID: "3"})
		return tickets, nil
	})

	schema, err := builder.Build()

	if err != nil {
		panic(err)
	}

	// Query
	query := `
		{
			ticket { title, id }
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON) // {"data":{"hello":"world"}}

}
