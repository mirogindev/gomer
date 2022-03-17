package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/mirogindev/gomer/test_uttils"
	"log"
	"testing"
)

func BuildTestSchema() (graphql.Schema, error) {
	builder := test_uttils.CreateTestSchema()

	schema, err := builder.Build()
	return schema, err
}

//func TestRunWeb(t *testing.T) {
//	schema, err := BuildTestSchema()
//	if err != nil {
//		panic(err)
//	}
//
//	if err != nil {
//		log.Fatalf("failed to create new schema, error: %v", err)
//	}
//
//	h := handler.New(&handler.Config{
//		Schema:     &schema,
//		Pretty:     true,
//		GraphiQL:   false,
//		Playground: true,
//	})
//
//	sh := gqbuilder.GetSubscriptionHandler(schema)
//
//	http.Handle("/graphql", h)
//
//	http.HandleFunc("/subscriptions", sh.SubscriptionsHandlerFunc)
//	log.Fatal(http.ListenAndServe(":8089", nil))
//}

func TestQuery(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		{
			ticket_without_arguments { title }
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

func TestQueryWithArgs(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		{
			ticket(limit: 10, filter:{ title: { eq: "ddd" } } ) { title }
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

func TestMutation(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		mutation {
			ticket_insert (input:{title:"t1", numbers:[1,2],numbers_required:[4,5]}) { title }
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

func TestMutationDecimal(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		mutation {
			ticket_insert (input:{decimal:"0.1112"}) { title }
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

func TestMutationMany(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		mutation {
			ticket_insert_many (input:[{title:"title1"},{ title:"title2"}]) { title }
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
			ticket(limit: 15, offset: 10, filter:{ title: { eq :"ddd"}, id: { neq: 500 } } ) { title, tags( limit: 10) { title } }
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

func TestParseNestedArgs(t *testing.T) {

	schema, err := BuildTestSchema()

	if err != nil {
		panic(err)
	}

	query := `
		{
			ticket(limit: 15, offset: 17, filter:{ title:{eq: "ddd" } } ) { title, tags(limit:5,offset:6){ title } }
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

func pretty(x interface{}) string {
	got, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(got)
}
