package gqbuilder

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"gogql/models"
	"reflect"
)

var scalarsMap = map[string]*graphql.Scalar{
	"string":   graphql.String,
	"int":      graphql.Int,
	"int64":    graphql.Int,
	"float":    graphql.Float,
	"float64":  graphql.Float,
	"datetime": graphql.DateTime,
	"bool":     graphql.Boolean,
}

func isScalar(t reflect.Type) (*graphql.Scalar, bool) {
	n := t.Name()
	if v, ok := scalarsMap[n]; ok {
		return v, true
	}
	return nil, false
}

func getKey(t reflect.Type) string {
	pkg := t.PkgPath
	nk := t.Name()

	return fmt.Sprintf("%s/%s", pkg(), nk)
}

func BuildTestSchema() (graphql.Schema, error) {
	builder := SchemaBuilder{}

	//ticket := builder.Object("Ticket", models.Ticket{})

	queryObj := builder.Query()

	queryObj.FieldFunc("ticket", func(ctx context.Context, args struct {
		Filter *models.TicketFilterInput
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
	return schema, err
}
