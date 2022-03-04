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

	queryObj := builder.Query()

	queryObj.FieldFunc("ticket", func(ctx context.Context, rgs struct {
		Filter *models.TagFilterInput
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

func Test1(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	// Query
	//query := `
	//	{
	//		ticket(limit: 15, offset: 10) { title }
	//	}`

	query := `
		{
			ticket(limit: 15, offset: 10, filter:{ title:"ddd" } ) { title }
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON)

}

func TestWithRelationQuery(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	// Query
	//query := `
	//	{
	//		ticket(limit: 15, offset: 10) { title }
	//	}`

	query := `
		{
			ticket(limit: 15, offset: 10, filter:{ title:"ddd" } ) { title, tags( limit: 10) { title } }
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON)

}
