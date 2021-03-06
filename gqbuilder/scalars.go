package gqbuilder

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/shopspring/decimal"
	log "github.com/sirupsen/logrus"
	"strconv"
)

//var JsonField = graphql.NewScalar(graphql.ScalarConfig{
//	Name:        "JSON",
//	Description: ``,
//	Serialize: func(value interface{}) interface{} {
//		switch value := value.(type) {
//		case *datatypes.JSON:
//			return *value
//		case datatypes.JSON:
//			return value
//		default:
//			panic("Value is not JSON")
//		}
//
//		return nil
//	},
//	// parseValue: gets invoked to parse client input that was passed through variables.
//	// value is plain type
//	ParseValue: func(value interface{}) interface{} {
//		switch value := value.(type) {
//		case *datatypes.JSON:
//			return *value
//		case datatypes.JSON:
//			return value
//		default:
//			panic("Value is not JSON")
//		}
//
//		return nil
//	},
//	// parseLiteral: gets invoked to parse client input that was passed inline in the query.
//	// value is ast.Value
//	ParseLiteral: func(valueAST ast.Value) interface{} {
//		switch valueAST := valueAST.(type) {
//		case *ast.StringValue:
//			res := datatypes.JSON{}
//			bytes := []byte(valueAST.Value)
//			err := res.UnmarshalJSON(bytes)
//			if err != nil {
//				panic(err)
//			}
//
//			return res
//		}
//		return nil
//	},
//})

var Int64Scalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "int64",
	Description: ``,
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case *int64:
			return *value
		case int64:
			return value
		default:
			log.Errorf("Value is not int64, actial type is %v", value)
		}

		return nil
	},
	// parseValue: gets invoked to parse client input that was passed through variables.
	// value is plain type
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case int64:
			return value
		case float64:
			return int64(value)
		case string:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				log.Errorf("Cannot convert %v to int64, %s", value, err)
			}
			return i
		default:
			log.Errorf("Value is not int64, actial type is %v", value)
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
				log.Error(err)
			}

			return i
		}
		return nil
	},
})

var DecimalScalar = graphql.NewScalar(graphql.ScalarConfig{
	Name:        "Decimal",
	Description: ``,
	Serialize: func(value interface{}) interface{} {
		switch value := value.(type) {
		case *decimal.Decimal:
			return *value
		case decimal.Decimal:
			return value
		default:
			log.Errorf("Value is not decimal, actial type is %v", value)
		}

		return nil
	},
	// parseValue: gets invoked to parse client input that was passed through variables.
	// value is plain type
	ParseValue: func(value interface{}) interface{} {
		switch value := value.(type) {
		case *decimal.Decimal:
			return *value
		case decimal.Decimal:
			return value
		case string:
			i, err := decimal.NewFromString(value)
			if err != nil {
				log.Errorf("Cannot convert %v to decimal, %s", value, err)
			}
			return i
		default:
			log.Errorf("Value is not decimal, actial type is %v", value)
		}

		return nil
	},
	// parseLiteral: gets invoked to parse client input that was passed inline in the query.
	// value is ast.Value
	ParseLiteral: func(valueAST ast.Value) interface{} {
		switch valueAST := valueAST.(type) {
		case *ast.StringValue:
			i, err := decimal.NewFromString(valueAST.Value)
			if err != nil {
				log.Error(err)
			}

			return i
		}
		return nil
	},
})
