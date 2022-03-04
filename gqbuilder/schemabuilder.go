package gqbuilder

import (
	"context"
	"github.com/graphql-go/graphql"
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

type schemaBuilder struct {
	objects      map[string]*Object
	builtObjects map[string]graphql.Output
	builtInputs  map[string]graphql.Input
}

func GetBuilder() *schemaBuilder {
	return &schemaBuilder{}
}

func (s *schemaBuilder) Query() *Object {
	name := Query
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: query{},
	}
	return s.objects[name]
}

func (s *schemaBuilder) Mutation() *Object {
	name := Mutation
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: mutation{},
	}
	return s.objects[name]
}

func (s *schemaBuilder) Object(name string, obj interface{}) *Object {
	s.checkObjects(name)

	s.objects[name] = &Object{
		Name: name,
		Type: obj,
	}
	return s.objects[name]
}

func (s *schemaBuilder) checkObjects(name string) {
	if s.objects == nil {
		s.objects = make(map[string]*Object)
	}
	if _, ok := s.objects[name]; ok {
		log.Panicf("Func with name %s aready exists", name)
	}
}

func (s *schemaBuilder) buildObjectFields(t reflect.Type) graphql.Fields {
	fields := graphql.Fields{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n, fo := s.buildField(f)
		fields[n] = fo
	}

	return fields
}

func (s *schemaBuilder) buildObject(name string, t reflect.Type) graphql.Output {
	fields := s.buildObjectFields(t)

	return graphql.NewObject(graphql.ObjectConfig{
		Name:   name,
		Fields: fields,
	})
}

func (s *schemaBuilder) buildInputObject(name string, t reflect.Type) graphql.Input {
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

func (s *schemaBuilder) buildFieldConfigArgument(t reflect.Type) graphql.FieldConfigArgument {
	fields := graphql.FieldConfigArgument{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		n, fo := s.buildArgumentConfig(f)
		fields[n] = fo
	}

	return fields
}

func (s *schemaBuilder) buildArgumentConfig(reflectedType reflect.StructField) (string, *graphql.ArgumentConfig) {
	n := getFieldName(reflectedType.Name)
	gqType := s.getGqInput(reflectedType.Type, true)

	field := graphql.ArgumentConfig{
		Type: gqType,
	}
	return n, &field
}

func (s *schemaBuilder) buildInputField(reflectedType reflect.StructField) (string, *graphql.InputObjectFieldConfig) {
	n := getFieldName(reflectedType.Name)
	gqType := s.getGqInput(reflectedType.Type, true)

	field := graphql.InputObjectFieldConfig{
		Type: gqType,
	}
	return n, &field
}

func (s *schemaBuilder) buildField(reflectedType reflect.StructField) (string, *graphql.Field) {
	n := getFieldName(reflectedType.Name)
	gqType := s.getGqOutput(reflectedType.Type, true)

	field := graphql.Field{
		Name: n,
		Type: gqType,
	}
	return n, &field
}

func (s *schemaBuilder) getGqInput(reflectedType reflect.Type, isRequired bool) graphql.Input {
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

func (s *schemaBuilder) getGqOutput(reflectedType reflect.Type, isRequired bool) graphql.Output {
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

func (s *schemaBuilder) buildObjects() error {
	for n, v := range s.objects {
		t := reflect.TypeOf(v.Type)
		methodFields := s.buildMethods(v.Methods)
		objectFields := s.buildObjectFields(t)
		fields := mergeFields(methodFields, objectFields)
		obj := graphql.NewObject(graphql.ObjectConfig{Name: n, Fields: fields})
		key := getKey(t)
		s.builtObjects[key] = obj
	}
	return nil
}

func (s *schemaBuilder) buildQuery() *graphql.Object {
	if qf, ok := s.objects[Query]; ok {
		fields := s.buildMethods(qf.Methods)
		rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: "RootQuery", Fields: fields})

		return rootQuery
	}
	log.Fatalln("Query object is not found")
	return nil
}

func (s *schemaBuilder) buildMutation() *graphql.Object {
	if qf, ok := s.objects[Mutation]; ok {
		fields := s.buildMethods(qf.Methods)
		rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: "RootMutation", Fields: fields})

		return rootQuery
	}
	log.Fatalln("Mutation object is not found")
	return nil
}

func (s *schemaBuilder) buildMethods(methods map[string]*Method) graphql.Fields {
	fields := graphql.Fields{}
	for n, v := range methods {
		ro := s.getResolverOutput(v.Fn)
		args := s.getResolverArgs(v.Fn)
		fun := s.getFunc(v.Fn)
		fields[n] = &graphql.Field{
			Args: args,
			Type: ro,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				in := make([]reflect.Value, fun.Type().NumIn())
				if p.Context == nil {
					in[0] = reflect.New(fun.Type().In(0)).Elem()
				} else {
					in[0] = reflect.ValueOf(p.Context)
				}

				if p.Args == nil {
					//argType,_ := getArgs(fun.Type())
					//args := reflect.New(argType).Elem()
					////argObj := reflect.TypeOf(args)
					//////for i := 0; i < argObj.NumField(); i++ {
					//////	f := argObj.Field(i)
					//////}

				} else {
					argType, pos := getArgs(fun.Type())
					if _, ok := p.Source.(map[string]interface{}); !ok {
						in[pos-1] = reflect.ValueOf(p.Source)
					}
					args := ReflectStruct(argType, p.Args)
					in[pos] = args
				}

				result := fun.Call(in)

				var respData interface{}
				var err error
				if result[0].Interface() != nil {
					respData = result[0].Interface()
				} else {
					respData = reflect.New(fun.Type().Out(0)).Elem()
				}

				if result[1].Interface() != nil {
					err = result[1].Interface().(error)
				}

				return respData, err
			},
		}
	}
	return fields
}

func ReflectStruct(t reflect.Type, params map[string]interface{}) reflect.Value {
	val := reflect.New(t).Elem()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() == reflect.Ptr {
			if f.Type.Elem().Kind() == reflect.Struct {
				fName := getFieldName(f.Name)
				if np, ok := params[fName]; ok {
					pr := reflect.New(f.Type.Elem())
					pr.Elem().Set(ReflectStruct(f.Type.Elem(), np.(map[string]interface{})))
					val.Field(i).Set(pr)
				}
			} else if f.Type.Elem().Kind() == reflect.Slice {
				elemSlice := reflect.MakeSlice(reflect.SliceOf(f.Type.Elem().Elem()), 0, 5)
				fName := getFieldName(f.Name)
				if f.Type.Elem().Elem().Kind() == reflect.Ptr {
					if np, ok := params[fName]; ok {
						pr := reflect.New(f.Type.Elem())
						for _, v := range np.([]interface{}) {
							item := ReflectStruct(f.Type.Elem().Elem().Elem(), v.(map[string]interface{}))
							ipr := reflect.New(f.Type.Elem().Elem().Elem())
							ipr.Elem().Set(item)
							elemSlice = reflect.Append(elemSlice, ipr)
						}
						pr.Elem().Set(elemSlice)
						val.Field(i).Set(pr)
					}
				} else {
					if np, ok := params[fName]; ok {
						pr := reflect.New(f.Type.Elem())
						for _, v := range np.([]interface{}) {
							item := ReflectStruct(f.Type.Elem().Elem(), v.(map[string]interface{}))
							elemSlice = reflect.Append(elemSlice, item)
						}
						pr.Elem().Set(elemSlice)
						val.Field(i).Set(pr)
					}
				}
			} else {
				pr := reflect.New(f.Type.Elem())
				pr.Elem().Set(reflectField(f.Name, f.Type.Elem(), params))
				val.Field(i).Set(pr)
			}
		} else if f.Type.Kind() == reflect.Struct {
			fName := getFieldName(f.Name)
			if np, ok := params[fName]; ok {
				val.Field(i).Set(ReflectStruct(f.Type, np.(map[string]interface{})))
			}
		} else if f.Type.Kind() == reflect.Slice {
			elemSlice := reflect.MakeSlice(reflect.SliceOf(f.Type.Elem()), 0, 5)
			fName := getFieldName(f.Name)
			if f.Type.Elem().Kind() == reflect.Ptr {
				if np, ok := params[fName]; ok {
					for _, v := range np.([]interface{}) {
						item := ReflectStruct(f.Type.Elem().Elem(), v.(map[string]interface{}))
						pr := reflect.New(f.Type.Elem().Elem())
						pr.Elem().Set(item)
						elemSlice = reflect.Append(elemSlice, pr)
					}
					val.Field(i).Set(elemSlice)
				}
			} else {
				if np, ok := params[fName]; ok {
					for _, v := range np.([]interface{}) {
						item := ReflectStruct(f.Type.Elem(), v.(map[string]interface{}))
						elemSlice = reflect.Append(elemSlice, item)
					}
					val.Field(i).Set(elemSlice)
				}
			}
		} else {
			val.Field(i).Set(reflectField(f.Name, f.Type, params))
		}
	}
	return val
}

func reflectField(name string, f reflect.Type, params map[string]interface{}) reflect.Value {
	val := params[getFieldName(name)]
	if val == nil {
		return reflect.New(f).Elem()
	}
	return reflect.ValueOf(val)
}

//func reflectValue(name string, t reflect.Type, v map[string]interface{}) reflect.Value {
//	log.Println(name)
//	val := v[strcase.ToSnake(name)]
//	if val == nil {
//		return reflect.New(t)
//	}
//	if t.Kind() == reflect.Ptr {
//		pr := reflect.New(t.Elem())
//		pr.Elem().Set(reflectValue(name, t.Elem(), v))
//		return pr
//	} else if t.Kind() == reflect.Struct {
//		se := reflect.New(t).Elem()
//		for i := 0; i < t.NumField(); i++ {
//			f := t.Field(i)
//			val2 := reflectValue(f.Name, f.Type, v[strcase.ToSnake(name)].(map[string]interface{}))
//			se.Field(i).Set(val2)
//		}
//
//		return se
//	}
//	//tn := t.Name()
//	//tk := t.Kind()
//	//log.Println(tn, tk)
//	//	pr := reflect.New(t)
//	//pr.Set(reflect.ValueOf(val))
//	return reflect.ValueOf(val)
//}

func (s *schemaBuilder) getResolverOutput(fn interface{}) graphql.Output {
	rf := reflect.TypeOf(fn).Out(0)
	return s.getGqOutput(rf, true)
}

func (s *schemaBuilder) getResolverArgs(fn interface{}) graphql.FieldConfigArgument {
	//t := reflect.TypeOf(fn)
	//if t.NumIn() == 2 {
	//	rf := reflect.TypeOf(fn).In(1)
	//	return s.buildFieldConfigArgument(rf)
	//} else {
	//	rf := reflect.TypeOf(fn).In(2)
	//	return s.buildFieldConfigArgument(rf)
	//}
	args, _ := getArgs(reflect.TypeOf(fn))
	return s.buildFieldConfigArgument(args)
}

func (s *schemaBuilder) getFunc(fn interface{}) reflect.Value {
	rf := reflect.ValueOf(fn)
	return rf
}

func (s *schemaBuilder) Build() (graphql.Schema, error) {
	s.buildObjects()
	mutation := s.buildMutation()
	query := s.buildQuery()

	schemaConfig := graphql.SchemaConfig{Query: query, Mutation: mutation}
	schema, err := graphql.NewSchema(schemaConfig)
	return schema, err
}

type FieldResolveFn func() (interface{}, error)
