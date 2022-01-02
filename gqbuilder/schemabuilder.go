package gqbuilder

import (
	"fmt"
	"github.com/graphql-go/graphql"
)

type query struct{}
type mutation struct{}

type SchemaBuilder struct {
	objects map[string]*Object
}

func (s *SchemaBuilder) Query() *Object {
	name := "Query"
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: query{},
	}
	return s.objects[name]
}

func (s *SchemaBuilder) Mutation() *Object {
	name := "Mutation"
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: mutation{},
	}
	return s.objects[name]
}

func (s *SchemaBuilder) Object(name string, obj interface{}) *Object {
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: obj,
	}
	return s.objects[name]
}

func (s *SchemaBuilder) checkObjects(name string) {
	if s.objects == nil {
		s.objects = make(map[string]*Object)
	}
	if s.objects[name] != nil {
		panic(fmt.Sprintf("Object with name %s aready exists", name))
	}
}

func (s *SchemaBuilder) Build() (graphql.Schema, error) {
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: graphql.String,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	return schema, err
}

type FieldResolveFn func() (interface{}, error)
