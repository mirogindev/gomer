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
	"time"
)

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

	mutationObj.FieldResolver("ticket_insert", func(ctx context.Context, rgs struct {
		input *models.TicketInsertInput
	}) (*models.Ticket, error) {
		var tags []*models.Tag
		tags = append(tags, &models.Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &models.Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &models.Tag{Title: "Tag3", ID: "3"})

		return &models.Ticket{Title: "Ticket1", ID: "1", Tags: tags}, nil
	})

	queryObj := builder.Query()

	queryObj.FieldResolver("ticket", func(ctx context.Context, args struct {
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

	subObj := builder.Subscription()

	subObj.FieldSubscription("test_sub", func(ctx context.Context, c chan *models.Ticket, args struct {
		Filter *models.TagFilterInput
		Order  *models.TagOrderInput
		Limit  *int
		Offset *int
	}) (chan *models.Ticket, error) {
		go func() {
			var i int

			for {
				i++

				ticket := models.Ticket{ID: fmt.Sprintf("%d", i)}

				select {
				case <-ctx.Done():
					log.Println("[RootSubscription] [Subscribe] subscription canceled")
					close(c)
					return
				default:
					c <- &ticket
				}

				time.Sleep(2000 * time.Millisecond)

				if i == 1000 {
					close(c)
					return
				}
			}
		}()
		return c, nil
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

func TestSubscription(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
				subscription {
					test_sub
				}
			`
	c := graphql.Subscribe(graphql.Params{
		Context:       ctx,
		OperationName: "",
		RequestString: query,
		Schema:        schema,
	})

	var results []*graphql.Result
	for res := range c {
		t.Log(pretty(res))
		results = append(results, res)
	}

}

func pretty(x interface{}) string {
	got, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(got)
}
