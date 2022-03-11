package gqbuilder

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	"github.com/mirogindev/gomer/models"
	"reflect"
)

var scalarsMap = map[string]*graphql.Scalar{
	"string":   graphql.String,
	"int":      graphql.Int,
	"int64":    graphql.Int,
	"float64":  graphql.Float,
	"float32":  graphql.Float,
	"datetime": graphql.DateTime,
	"Time":     graphql.DateTime,
	"Decimal":  graphql.String,
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

type Args struct {
	Filter *models.TicketFilterInput
	Limit  int
	Offset *int
}

func getFieldName(name string) string {
	return strcase.ToSnake(name)
}

func mergeFields(methodFields, objectFields graphql.Fields) graphql.Fields {
	for k, v := range methodFields {
		objectFields[k] = v
	}
	return objectFields
}

func getArgs(fun reflect.Type) (reflect.Type, int) {
	if fun.NumIn() == 3 {
		pos := 2
		return fun.In(pos), pos
	} else {
		pos := 1
		return fun.In(pos), pos
	}
}
