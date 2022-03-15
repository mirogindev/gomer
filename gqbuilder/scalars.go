package gqbuilder

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"strconv"
)

var Int64Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Int64",
	Description: ``,
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case *int64:
			return *value
		case int64:
			return value
		default:
			panic("Value is not int64")
		}

		return nil
	},
	// parseValue: gets invoked to parse client input that was passed through variables.
	// value is plain type
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case *int64:
			return *value
		case int64:
			return value
		default:
			panic("Value is not int64")
		}

		return nil
	},
	// parseLiteral: gets invoked to parse client input that was passed inline in the query.
	// value is ast.Value
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.IntValue:
			i, err := strconv.ParseInt(valueAST.Value, 10, 64)
			if err != nil {
				panic(err)
			}

			return i
		}
		return nil
	},
})
