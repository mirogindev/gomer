package gqbuilder

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/iancoleman/strcase"
	"github.com/mirogindev/gomer/models"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"math"
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

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getKey(t reflect.Type) string {
	//pkg := t.PkgPath
	nk := t.Name()

	if nk == "" {
		return t.String()
		//	return fmt.Sprintf("Args%v", hash(t.String()))
	}

	return fmt.Sprintf("%s", nk)
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

func getArgumentValues(
	argDefs []*graphql.Argument, argASTs []*ast.Argument,
	variableValues map[string]interface{}) map[string]interface{} {

	argASTMap := map[string]*ast.Argument{}
	for _, argAST := range argASTs {
		if argAST.Name != nil {
			argASTMap[argAST.Name.Value] = argAST
		}
	}
	results := map[string]interface{}{}
	for _, argDef := range argDefs {
		var (
			tmp   interface{}
			value ast.Value
		)
		if tmpValue, ok := argASTMap[argDef.PrivateName]; ok {
			value = tmpValue.Value
		}
		if tmp = valueFromAST(value, argDef.Type, variableValues); isNullish(tmp) {
			tmp = argDef.DefaultValue
		}
		if !isNullish(tmp) {
			results[argDef.PrivateName] = tmp
		}
	}
	return results
}

func valueFromAST(valueAST ast.Value, ttype graphql.Input, variables map[string]interface{}) interface{} {
	if valueAST == nil {
		return nil
	}
	// precedence: value > type
	if valueAST, ok := valueAST.(*ast.Variable); ok {
		if valueAST.Name == nil || variables == nil {
			return nil
		}
		// Note: we're not doing any checking that this variable is correct. We're
		// assuming that this query has been validated and the variable usage here
		// is of the correct type.
		return variables[valueAST.Name.Value]
	}
	switch ttype := ttype.(type) {
	case *graphql.NonNull:
		return valueFromAST(valueAST, ttype.OfType, variables)
	case *graphql.List:
		values := []interface{}{}
		if valueAST, ok := valueAST.(*ast.ListValue); ok {
			for _, itemAST := range valueAST.Values {
				values = append(values, valueFromAST(itemAST, ttype.OfType, variables))
			}
			return values
		}
		return append(values, valueFromAST(valueAST, ttype.OfType, variables))
	case *graphql.InputObject:
		var (
			ok bool
			ov *ast.ObjectValue
			of *ast.ObjectField
		)
		if ov, ok = valueAST.(*ast.ObjectValue); !ok {
			return nil
		}
		fieldASTs := map[string]*ast.ObjectField{}
		for _, of = range ov.Fields {
			if of == nil || of.Name == nil {
				continue
			}
			fieldASTs[of.Name.Value] = of
		}
		obj := map[string]interface{}{}
		for name, field := range ttype.Fields() {
			var value interface{}
			if of, ok = fieldASTs[name]; ok {
				value = valueFromAST(of.Value, field.Type, variables)
			} else {
				value = field.DefaultValue
			}
			if !isNullish(value) {
				obj[name] = value
			}
		}
		return obj
	case *graphql.Scalar:
		return ttype.ParseLiteral(valueAST)
	case *graphql.Enum:
		return ttype.ParseLiteral(valueAST)
	}

	return nil
}

func isNullish(src interface{}) bool {
	if src == nil {
		return true
	}
	value := reflect.ValueOf(src)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return true
		}
		value = value.Elem()
	}
	switch value.Kind() {
	case reflect.String:
		// if src is ptr type and len(string)=0, it returns false
		if !value.IsValid() {
			return true
		}
	case reflect.Int:
		return math.IsNaN(float64(value.Int()))
	case reflect.Float32, reflect.Float64:
		return math.IsNaN(float64(value.Float()))
	}
	return false
}

func getFieldDef(schema graphql.Schema, parentType *graphql.Object, fieldName string) *graphql.FieldDefinition {

	if parentType == nil {
		return nil
	}

	if fieldName == graphql.SchemaMetaFieldDef.Name &&
		schema.QueryType() == parentType {
		return graphql.SchemaMetaFieldDef
	}
	if fieldName == graphql.TypeMetaFieldDef.Name &&
		schema.QueryType() == parentType {
		return graphql.TypeMetaFieldDef
	}
	if fieldName == graphql.TypeNameMetaFieldDef.Name {
		return graphql.TypeNameMetaFieldDef
	}
	return parentType.Fields()[fieldName]
}

func parseSelection(f ast.Selection, parentType *graphql.Object, p graphql.ResolveParams, argsMap map[string]map[string]interface{}) *Selection {
	switch v := f.(type) {
	case *ast.Field:
		var sel *Selection
		var args interface{}
		fieldDef := getFieldDef(p.Info.Schema, parentType, v.Name.Value)

		sel = &Selection{
			Name: fieldDef.Name,
		}

		if v.SelectionSet == nil {

			return sel
		}

		argsObject := argsMap[parentType.Name()][fieldDef.Name]
		if argsObject != nil {
			parsedArgs := getArgumentValues(fieldDef.Args, v.Arguments, p.Info.VariableValues)
			args = ReflectStruct(reflect.TypeOf(argsObject), parsedArgs).Interface()
		}

		selections := make([]*Selection, 0)

		for _, s := range v.SelectionSet.Selections {
			fieldObject := getFieldObject(fieldDef.Type)
			log.Println(fieldObject)
			ps := parseSelection(s, fieldObject, p, argsMap)
			selections = append(selections, ps)
		}

		sel.Args = args
		sel.SelectionSet = &SelectionSet{
			Selections: selections,
		}

		return sel
	default:
		log.Panicf("Invalid type %s", v)
	}
	return nil
}

func ParseSelections(p graphql.ResolveParams, argsMap map[string]map[string]interface{}) []*Selection {
	selections := make([]*Selection, 0)
	od := p.Info.Operation.(*ast.OperationDefinition)
	s := od.GetSelectionSet().Selections
	for _, f := range s {
		selections = append(selections, parseSelection(f, p.Info.ParentType.(*graphql.Object), p, argsMap))
	}
	return selections
}

func getFieldObject(f graphql.Type) *graphql.Object {
	switch v := f.(type) {
	case *graphql.List:
		return getFieldObject(v.OfType)
	case *graphql.NonNull:
		return getFieldObject(v.OfType)
	case *graphql.Object:
		return v
	}
	log.Panicf("Cannot get field object type")
	return nil
}

func getActualTypeRecursive(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Ptr:
		return getActualTypeRecursive(t.Elem())
	case reflect.Slice:
		return getActualTypeRecursive(t.Elem())

	case reflect.Struct:
		return t
	}
	return t

}
