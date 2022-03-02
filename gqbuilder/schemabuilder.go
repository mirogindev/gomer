package gqbuilder

import (
	"context"
	"github.com/graphql-go/graphql"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"reflect"
)

const (
	Query    = "Query"
	Mutation = "Mutation"
)

type HandlerFn func(ctx context.Context, args interface{}) (interface{}, error)

type query struct{}
type mutation struct{}

type SchemaBuilder struct {
	objects      map[string]*Object
	builtObjects map[string]graphql.Output
	builtInputs  map[string]graphql.Input
}

func (s *SchemaBuilder) Query() *Object {
	name := Query
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: query{},
	}
	return s.objects[name]
}

func (s *SchemaBuilder) Mutation() *Object {
	name := Mutation
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
	if _, ok := s.objects[name]; ok {
		log.Panicf("Func with name %s aready exists", name)
	}
}

func (s *SchemaBuilder) buildObject(name string, t reflect.Type) graphql.Output {
	fields := graphql.Fields{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n, fo := s.buildField(f)
		fields[n] = fo
	}

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	})
}

func (s *SchemaBuilder) buildInputObject(name string, t reflect.Type) graphql.Input {
	fields := graphql.InputObjectConfigFieldMap{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n, fo := s.buildInputField(f)
		fields[n] = fo
	}

	return graphql.NewInputObject(graphql.InputObjectConfig{
		Name:   name,
		Fields: fields,
	})
}

func (s *SchemaBuilder) buildFieldConfigArgument(t reflect.Type) graphql.FieldConfigArgument {
	fields := graphql.FieldConfigArgument{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n, fo := s.buildArgumentConfig(f)
		fields[n] = fo
	}

	return fields
}

func (s *SchemaBuilder) buildArgumentConfig(reflectedType reflect.StructField) (string, *graphql.ArgumentConfig) {
	n := strcase.ToSnake(reflectedType.Name)
	gqType := s.getGqInput(reflectedType.Type, true)

	field := graphql.ArgumentConfig{
		Type: gqType,
	}
	return n, &field
}

func (s *SchemaBuilder) buildInputField(reflectedType reflect.StructField) (string, *graphql.InputObjectFieldConfig) {
	n := strcase.ToSnake(reflectedType.Name)
	gqType := s.getGqInput(reflectedType.Type, true)

	field := graphql.InputObjectFieldConfig{
		Type: gqType,
	}
	return n, &field
}

func (s *SchemaBuilder) buildField(reflectedType reflect.StructField) (string, *graphql.Field) {
	n := strcase.ToSnake(reflectedType.Name)
	gqType := s.getGqOutput(reflectedType.Type, true)

	field := graphql.Field{
		Name: n,
		Type: gqType,
	}
	return n, &field
}

func (s *SchemaBuilder) getGqInput(reflectedType reflect.Type, isRequired bool) graphql.Input {
	if s.builtInputs == nil {
		s.builtInputs = make(map[string]graphql.Input)
	}

	if reflectedType.Kind() == reflect.Ptr {
		return s.getGqInput(reflectedType.Elem(), false)
	} else if reflectedType.Kind() == reflect.Slice {
		obj := graphql.NewList(s.getGqInput(reflectedType.Elem(), false))
		if isRequired {
			return graphql.NewNonNull(obj)
		}
		return obj

	} else if reflectedType.Kind() == reflect.Struct {
		key := getKey(reflectedType)
		if v, ok := s.builtInputs[key]; ok {
			return v
		}

		bo := s.buildInputObject(reflectedType.Name(), reflectedType)
		s.builtInputs[key] = bo

		return bo
	}
	v, ok := isScalar(reflectedType)
	if ok {
		if isRequired {
			return graphql.NewNonNull(v)
		}
		return v
	}
	return nil
}

func (s *SchemaBuilder) getGqOutput(reflectedType reflect.Type, isRequired bool) graphql.Output {
	if s.builtObjects == nil {
		s.builtObjects = make(map[string]graphql.Output)
	}

	if reflectedType.Kind() == reflect.Ptr {
		return s.getGqOutput(reflectedType.Elem(), false)
	} else if reflectedType.Kind() == reflect.Slice {
		obj := graphql.NewList(s.getGqOutput(reflectedType.Elem(), false))
		if isRequired {
			return graphql.NewNonNull(obj)
		}
		return obj

	} else if reflectedType.Kind() == reflect.Struct {
		key := getKey(reflectedType)
		if v, ok := s.builtObjects[key]; ok {
			return v
		}

		bo := s.buildObject(reflectedType.Name(), reflectedType)
		s.builtObjects[key] = bo

		return bo
	}
	v, ok := isScalar(reflectedType)
	if ok {
		if isRequired {
			return graphql.NewNonNull(v)
		}
		return v
	}
	return nil
}

func (s *SchemaBuilder) buildQuery() *graphql.Object {
	fields := graphql.Fields{}
	if queryFields, ok := s.objects[Query]; ok {
		methods := queryFields.Methods
		for n, v := range methods {
			ro := s.getResolverOutput(v.Fn)
			args := s.getResolverArgs(v.Fn)
			fields[n] = &graphql.Field{
				Args: args,
				Type: ro,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			}
		}
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})

	return rootQuery
}

func (s *SchemaBuilder) getResolverOutput(fn interface{}) graphql.Output {
	rf := reflect.TypeOf(fn).Out(0)
	return s.getGqOutput(rf, true)
}

func (s *SchemaBuilder) getResolverArgs(fn interface{}) graphql.FieldConfigArgument {
	rf := reflect.TypeOf(fn).In(1)
	return s.buildFieldConfigArgument(rf)
}

func (s *SchemaBuilder) Build() (graphql.Schema, error) {
	//	s.buildObjects()
	query := s.buildQuery()
	//fields.
	//fields := graphql.Fields{
	//	"hello": &graphql.Field{
	//		Type: graphql.String,
	//		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
	//			return "world", nil
	//		},
	//	},
	//}

	schemaConfig := graphql.SchemaConfig{Query: query}
	schema, err := graphql.NewSchema(schemaConfig)
	return schema, err
}

type FieldResolveFn func() (interface{}, error)
