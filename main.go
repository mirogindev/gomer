package main

import (
	"github.com/graphql-go/handler"
	"gogql/gqbuilder"
	"net/http"
)

func main() {

	schema, err := gqbuilder.BuildTestSchema()

	if err != nil {
		panic(err)
	}

	h := handler.New(&handler.Config{
		Schema:     &schema,
		Pretty:     true,
		GraphiQL:   false,
		Playground: true,
	})

	http.Handle("/graphql", h)
	http.ListenAndServe(":8080", nil)

}
