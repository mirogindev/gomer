package main

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/handler"
	"log"
	"net/http"
)

func main() {

	humanType := graphql.NewObject(graphql.ObjectConfig{
		Name:        "Human",
		Description: "A humanoid creature in the Star Wars universe.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.NewNonNull(graphql.String),
				Description: "The id of the human.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the human.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					return "name some", nil
				},
			},
		},
	})

	// Schema
	fields := graphql.Fields{
		"hello": &graphql.Field{
			Type: humanType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return "world", nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)

	//// Query
	//query := `
	//	{
	//		hello
	//	}
	//`
	//params := graphql.Params{Schema: schema, RequestString: query}
	//r := graphql.Do(params)
	//if len(r.Errors) > 0 {
	//	log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	//}
	//rJSON, _ := json.Marshal(r)
	//fmt.Printf("%s \n", rJSON) // {"data":{"hello":"world"}}
}
