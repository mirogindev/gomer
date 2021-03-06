package gqbuilder

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
	"math"
	"reflect"
	"sort"
	"strings"
)

func hashFromArr(fMap map[string]string) string {
	arr := make([]string, 0)
	for k, _ := range fMap {
		arr = append(arr, k)
	}

	sort.Strings(arr)
	return fmt.Sprintf("%v", hash(strings.Join(arr, ",")))

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
	}

	return fmt.Sprintf("%s", nk)
}

func getKeyWithHash(t reflect.Type, ig map[string]string) string {
	key := getKey(t)
	hashStr := hashFromArr(ig)
	return fmt.Sprintf("%s%s", key, hashStr)
}

func getFieldName(name string) string {
	return strcase.ToSnake(name)
}

func getArgs(fun reflect.Type) (reflect.Type, int, bool) {
	if fun.NumIn() == 3 {
		pos := 2
		return fun.In(pos), pos, true
	} else if fun.NumIn() == 2 {
		pos := 1
		t := fun.In(pos)
		if t.Kind() == reflect.Chan {
			return nil, 0, false
		}

		return fun.In(pos), pos, true
	}
	return nil, 0, false
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
			args = ReflectStructRecursive(reflect.TypeOf(argsObject), parsedArgs).Interface()
		}

		selections := make([]*Selection, 0)

		for _, s := range v.SelectionSet.Selections {
			fieldObject := getFieldObject(fieldDef.Type)
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

func ReflectStructFieldRecursive(fName string, t reflect.Type, param interface{}) reflect.Value {
	v := reflect.New(t).Elem()

	switch t.Kind() {
	case reflect.Ptr:
		log.Tracef("Reflect Ptr FieldName: %s, Type: %s ", fName, t.String())
		rs := ReflectStructFieldRecursive(fName, t.Elem(), param)
		ptr := reflect.New(t.Elem())
		ptr.Elem().Set(rs)
		v.Set(ptr)
	case reflect.Struct:
		log.Tracef("Reflect Struct FieldName: %s Type: %s", fName, t.String())

		if reflect.TypeOf(param) == t {
			v.Set(reflect.ValueOf(param))
		} else {
			rs := ReflectStructRecursive(t, param)
			v.Set(rs)
		}
	case reflect.Slice:
		log.Tracef("Reflect Struct FieldName: %s Type: %s", fName, t.String())
		if reflect.TypeOf(param) == t {
			v.Set(reflect.ValueOf(param))
		} else {
			slice := reflect.MakeSlice(t, 0, 5)
			for _, ai := range param.([]interface{}) {
				var item reflect.Value
				if n, ok := ai.(map[string]interface{}); ok {
					item = ReflectStructFieldRecursive(fName, t.Elem(), n)

				} else {
					ts := t.Elem().String()
					log.Traceln(ts)
					item = ReflectStructFieldRecursive(fName, t.Elem(), ai)
				}

				slice = reflect.Append(slice, item)
			}
			v.Set(slice)
		}
	default:
		log.Tracef("Reflect Default FieldName: %s Type: %s", fName, t.String())
		if param != nil {
			v.Set(reflect.ValueOf(param))
		}
	}
	log.Tracef("Reflect Return Value %s FieldName: %s Type: %s", v.Interface(), fName, t.String())
	return v
}

func ReflectStructRecursive(t reflect.Type, param interface{}) reflect.Value {
	val := reflect.New(t).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		fieldName := getFieldName(f.Name)
		if n, ok := param.(map[string]interface{}); ok {
			if np, ok := n[fieldName]; ok {
				rs := ReflectStructFieldRecursive(fieldName, f.Type, np)
				val.Field(i).Set(rs)
			}
		} else {
			rs := ReflectStructFieldRecursive(fieldName, f.Type, param)
			val.Field(i).Set(rs)
		}
	}

	return val
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

func findGomerTags(t reflect.StructField) GomerTags {
	gomerTags := GomerTags{}

	if tag, ok := t.Tag.Lookup("gomer"); ok {
		tr := strings.TrimSpace(tag)
		spl := strings.Split(tr, ";")
		for _, v := range spl {
			pspl := strings.Split(strings.TrimSpace(v), ":")
			if len(pspl) < 2 {
				log.Errorf("Invalid param %s", v)
			} else {
				gomerTags[pspl[0]] = pspl[1]
			}
		}
	}
	return gomerTags
}

func stringToMap(s string) map[string]string {
	arr := strings.Split(s, ",")
	m := make(map[string]string)

	for _, i := range arr {
		m[i] = i
	}
	return m
}

func MakeObjectNullable(output graphql.Output) graphql.Output {
	switch v := output.(type) {
	case *graphql.NonNull:
		return v.OfType
	}
	return output
}
