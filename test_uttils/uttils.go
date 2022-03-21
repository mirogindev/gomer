package test_uttils

import (
	"context"
	"fmt"
	"github.com/mirogindev/gomer/gqbuilder"
	log "github.com/sirupsen/logrus"
	"time"
)

func CreateTestSchema() *gqbuilder.SchemaBuilder {
	builder := gqbuilder.GetBuilder()
	builder.RegisterScalar("int64", gqbuilder.Int64Scalar)
	builder.RegisterScalar("Decimal", gqbuilder.DecimalScalar)
	//builder.RegisterScalar("JSON", gqbuilder.JsonField)

	ticket := builder.Object("Ticket", Ticket{})

	ticket.FieldResolver("tags", func(ctx context.Context, o *Ticket, args struct {
		Filter *TagFilterInput
		Offset *int
		Limit  *int
	}) ([]*Tag, error) {
		return o.Tags, nil
	})

	mutationObj := builder.Mutation()

	mutationObj.FieldResolver("ticket_insert", func(ctx context.Context, args struct {
		Input *TicketInsertInput
	}) (*Ticket, error) {
		var tags []*Tag
		tags = append(tags, &Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &Tag{Title: "Tag3", ID: "3"})

		return &Ticket{Title: "Ticket1", ID: "1", Tags: tags}, nil
	})

	mutationObj.FieldResolver("ticket_insert_many", func(ctx context.Context, args struct {
		Input []*TicketInsertInput
	}) ([]*Ticket, error) {
		var tickets []*Ticket
		var tags []*Tag
		tags = append(tags, &Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &Tag{Title: "Tag3", ID: "3"})

		tickets = append(tickets, &Ticket{Title: "Ticket1", ID: "1", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket2", ID: "2", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket3", ID: "3", Tags: tags})
		return tickets, nil
	})

	mutationObj.FieldResolver("ticket_update", func(ctx context.Context, args struct {
		Input *TicketInsertInput
	}) (*Ticket, error) {
		var tags []*Tag
		tags = append(tags, &Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &Tag{Title: "Tag3", ID: "3"})

		return &Ticket{Title: "Ticket1", ID: "1", Tags: tags}, nil
	})

	queryObj := builder.Query()

	queryObj.FieldResolver("ticket", func(ctx context.Context, args struct {
		Filter *TicketFilterInput
		Order  *TicketOrderInput
		Limit  *int
		Offset *int
	}) ([]*Ticket, error) {
		var tickets []*Ticket
		var tags []*Tag
		tags = append(tags, &Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &Tag{Title: "Tag3", ID: "3"})

		tickets = append(tickets, &Ticket{Title: "Ticket1", ID: "1", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket2", ID: "2", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket3", ID: "3", Tags: tags})
		return tickets, nil
	})

	queryObj.FieldResolver("ticket_without_args", func(ctx context.Context) ([]*Ticket, error) {
		var tickets []*Ticket
		var tags []*Tag
		tags = append(tags, &Tag{Title: "Tag1", ID: "1"})
		tags = append(tags, &Tag{Title: "Tag2", ID: "2"})
		tags = append(tags, &Tag{Title: "Tag3", ID: "3"})

		tickets = append(tickets, &Ticket{Title: "Ticket1", ID: "1", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket2", ID: "2", Tags: tags})
		tickets = append(tickets, &Ticket{Title: "Ticket3", ID: "3", Tags: tags})
		return tickets, nil
	})

	subObj := builder.Subscription()

	subObj.FieldSubscription("test_sub_without_args", Ticket{}, func(ctx context.Context, c chan interface{}) {
		var i int

		for {
			i++

			ticket := Ticket{ID: fmt.Sprintf("%d", i), Number: i}

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

	subObj.FieldSubscription("test_sub", Ticket{}, func(ctx context.Context, c chan interface{}, args struct {
		Filter *TicketFilterInput
		Order  *TagOrderInput
		Limit  *int
		Offset *int
	}) {
		var i int

		for {
			i++

			ticket := Ticket{ID: fmt.Sprintf("%d", i), Number: i}

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
	return builder
}
