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

//func (s *SchemaBuilder) buildObjects() {
//	s.builtObjects = make(map[string]graphql.Output)
//
//	for n, o := range s.objects {
//		if n == Query || n == Mutation {
//			continue
//		}
//
//		if o.Type == nil {
//			log.Panicf("Type for object %s is not set", o.Name)
//		}
//
//		t := reflect.TypeOf(o.Type)
//		s.builtObjects[n] = s.buildObject(n, t)
//	}
//
//	//fields := graphql.Fields{}
//	//if queryFields, ok := s.objects[Query]; ok {
//	//	methods := queryFields.Methods
//	//	for n, _ := range methods {
//	//		fields[n] = &graphql.Field{
//	//			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
//	//				return nil, nil
//	//			},
//	//		}
//	//	}
//	//}
//	//
//	//rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})
//
//}

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

func (s *SchemaBuilder) buildField(reflectedType reflect.StructField) (string, *graphql.Field) {
	n := strcase.ToSnake(reflectedType.Name)
	gqType := s.getGqType(reflectedType.Type, true)

	field := graphql.Field{
		Name: n,
		Type: gqType,
	}
	return n, &field
}

func (s *SchemaBuilder) getGqType(reflectedType reflect.Type, isRequired bool) graphql.Output {
	if reflectedType.Kind() == reflect.Ptr {
		return s.getGqType(reflectedType.Elem(), false)
	} else if reflectedType.Kind() == reflect.Slice {
		if isRequired {
			return graphql.NewNonNull(graphql.NewList(s.getGqType(reflectedType.Elem(), false)))
		}
		return graphql.NewList(s.getGqType(reflectedType.Elem(), false))
	} else if reflectedType.Kind() == reflect.Struct {
		key := getKey(reflectedType)
		if v, ok := s.builtObjects[key]; ok {
			return v
		}
		return s.buildObject(reflectedType.Name(), reflectedType)
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
			rf := reflect.TypeOf(v.Fn).Out(0)
			rt := s.getGqType(rf, true)
			fields[n] = &graphql.Field{
				Type: rt,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			}
		}
	}

	rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})

	return rootQuery
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
