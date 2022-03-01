package tests

import (
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"gogql/gqbuilder"
	"log"
	"testing"
)

func Test1(t *testing.T) {

	schema, err := gqbuilder.BuildTestSchema()

	if err != nil {
		panic(err)
	}

	// Query
	query := `
		{
			ticket { title }
		}
	`
	params := graphql.Params{Schema: schema, RequestString: query}
	r := graphql.Do(params)
	if len(r.Errors) > 0 {
		log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
	}
	rJSON, _ := json.Marshal(r)
	fmt.Printf("%s \n", rJSON) // {"data":{"hello":"world"}}

}
