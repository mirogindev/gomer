package gqbuilder

import (
	"context"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/mirogindev/gomer/logger"
	log "github.com/sirupsen/logrus"
	"reflect"
)

const (
	Query        = "Query"
	Mutation     = "Mutation"
	Subscription = "Subscription"
)

type HandlerFn func(ctx context.Context, args interface{}) (interface{}, error)

type query struct{}
type mutation struct{}
type subscription struct{}

const (
	INPUT_TYPE  = "INPUT_TYPE"
	OUTPUT_TYPE = "OUTPUT_TYPE"
)

var defaultScalarsMap = map[string]*graphql.Scalar{
	"string":   graphql.String,
	"int":      graphql.Int,
	"int64":    Int64Scalar,
	"Decimal":  DecimalScalar,
	"float64":  graphql.Float,
	"float32":  graphql.Float,
	"datetime": graphql.DateTime,
	"Time":     graphql.DateTime,
	"bool":     graphql.Boolean,
}

type SchemaBuilder struct {
	subscriptions  *SubscriptionObject
	scalars        map[string]*graphql.Scalar
	objects        map[string]GomerObject
	customObjects  map[string]GomerObject
	outputsToBuild map[string]reflect.Type
	inputsToBuild  map[string]reflect.Type
	argsMap        map[string]map[string]interface{}
	builtOutputs   map[string]graphql.Output
	builtInputs    map[string]graphql.Input
}

func GetBuilder() *SchemaBuilder {
	return &SchemaBuilder{}
}

func (s *SchemaBuilder) FindObjectsToBuild() (map[string]reflect.Type, map[string]reflect.Type) {
	if s.outputsToBuild == nil {
		s.outputsToBuild = make(map[string]reflect.Type)
	}
	if s.inputsToBuild == nil {
		s.inputsToBuild = make(map[string]reflect.Type)
	}

	for _, gm := range s.customObjects {
		v := gm.(*Object)
		t := reflect.TypeOf(v.Type)
		key := getKey(t)
		s.outputsToBuild[key] = t
		s.findMethodObjectsRecursive(v)
	}

	for _, gm := range s.objects {
		t := reflect.TypeOf(gm.GetType())
		key := getKey(t)
		s.outputsToBuild[key] = t
		s.findMethodObjectsRecursive(gm)

	}

	logger.GetLogger().Infof("Build objects tree success, Found %v inputs and %v outputs to build", len(s.inputsToBuild), len(s.outputsToBuild))
	return s.inputsToBuild, s.outputsToBuild
}

func (s *SchemaBuilder) CreateObjects() (map[string]graphql.Input, map[string]graphql.Output) {

	if s.builtInputs == nil {
		s.builtInputs = make(map[string]graphql.Input)
	}
	if s.builtOutputs == nil {
		s.builtOutputs = make(map[string]graphql.Output)
	}

	for n, v := range s.inputsToBuild {
		fields := graphql.InputObjectConfigFieldMap{}

		obj := graphql.NewInputObject(graphql.InputObjectConfig{
			Name:   n,
			Fields: fields,
		})
		key := getKey(v)
		s.builtInputs[key] = obj
	}

	for n, v := range s.outputsToBuild {
		if n == Query || n == Mutation {
			continue
		}
		fields := graphql.Fields{}

		obj := graphql.NewObject(graphql.ObjectConfig{
			Name:   v.Name(),
			Fields: fields,
		})
		s.builtOutputs[getKey(v)] = obj
	}
	return s.builtInputs, s.builtOutputs
}

func (s *SchemaBuilder) Query() *Object {
	name := Query
	s.checkObjects(name)
	obj := &Object{
		Name: name,
		Type: query{},
	}

	s.objects[name] = obj
	return obj
}

func (s *SchemaBuilder) Subscription() *SubscriptionObject {
	name := Subscription
	s.checkObjects(name)

	obj := &SubscriptionObject{
		Name: name,
		Type: subscription{},
	}

	s.objects[name] = obj
	return obj
}

func (s *SchemaBuilder) Mutation() *Object {
	name := Mutation
	s.checkObjects(name)

	obj := &Object{
		Name: name,
		Type: mutation{},
	}

	s.objects[name] = obj
	return obj
}

func (s *SchemaBuilder) Object(name string, objType interface{}) *Object {
	s.checkCustomObjects(name)

	obj := &Object{
		Name: name,
		Type: objType,
	}

	s.customObjects[name] = obj
	return obj
}

func (s *SchemaBuilder) checkCustomObjects(name string) {
	if s.customObjects == nil {
		s.customObjects = make(map[string]GomerObject)
	}
	if _, ok := s.customObjects[name]; ok {
		log.Panicf("Cutsom object with name %s aready exists", name)
	}
}

func (s *SchemaBuilder) checkObjects(name string) {
	if s.objects == nil {
		s.objects = make(map[string]GomerObject)
	}
	if _, ok := s.objects[name]; ok {
		log.Panicf("Resolver with name %s aready exists", name)
	}
}

func (s *SchemaBuilder) checkScalars(name string) {
	if s.scalars == nil {
		s.scalars = make(map[string]*graphql.Scalar)
	}
	if _, ok := s.scalars[name]; ok {
		log.Panicf("Scalar with name %s aready exists", name)
	}
}

func (s *SchemaBuilder) checkSubscriptions(name string) {
	if s.subscriptions == nil {
		s.subscriptions = &SubscriptionObject{}
	}
	if _, ok := s.subscriptions.Methods[name]; ok {
		log.Panicf("Subscription with name %s aready exists", name)
	}
}

func (s *SchemaBuilder) buildFieldConfigArgument(t reflect.Type) graphql.FieldConfigArgument {
	fields := graphql.FieldConfigArgument{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		io := s.getResolverInputObjectRecursive(f.Type)

		fields[getFieldName(f.Name)] = &graphql.ArgumentConfig{
			Type: io,
		}
	}

	return fields
}

func (s *SchemaBuilder) buildQuery() *graphql.Object {
	if qf, ok := s.objects[Query]; ok {
		fields := s.buildMethods(qf.(*Object))
		rootQuery := graphql.NewObject(graphql.ObjectConfig{Name: Query, Fields: fields})

		return rootQuery
	}
	log.Debug("Query object is not found")
	return nil
}

func (s *SchemaBuilder) buildMutation() *graphql.Object {
	if qf, ok := s.objects[Mutation]; ok {
		fields := s.buildMethods(qf.(*Object))
		rootMutation := graphql.NewObject(graphql.ObjectConfig{Name: Mutation, Fields: fields})

		return rootMutation
	}
	log.Debug("Mutation object is not found")
	return nil
}

func (s *SchemaBuilder) buildSubscription() *graphql.Object {
	if sf, ok := s.objects[Subscription]; ok {
		fields := s.buildSubscriptionMethods(sf.(*SubscriptionObject))
		rootSubscription := graphql.NewObject(graphql.ObjectConfig{Name: Subscription, Fields: fields})

		return rootSubscription
	}
	log.Debug("Subscription object is not found")
	return nil
}

func (s *SchemaBuilder) findMethodObjectsRecursive(gm GomerObject) {
	if o, ok := gm.(*Object); ok {
		for _, v := range o.Methods {

			ro := s.findResolverOutputObject(v.Fn)
			s.processObject(ro, OUTPUT_TYPE)

			ao := s.findResolverArgsObject(v.Fn)
			if ao == nil {
				continue
			}

			s.processObject(ao, INPUT_TYPE)
		}
	}

	if o, ok := gm.(*SubscriptionObject); ok {
		for _, v := range o.Methods {

			ro := s.findSubscriptionOutputObject(v.Output)
			s.processObject(ro, OUTPUT_TYPE)

			ao := s.findResolverArgsObject(v.Fn)

			if ao == nil {
				continue
			}
			s.processObject(ao, INPUT_TYPE)
		}
	}
}

func (s *SchemaBuilder) processObject(t reflect.Type, objType string) {
	if _, ok := s.isScalar(t); ok {
		return
	}
	key := getKey(t)
	if objType == INPUT_TYPE {
		s.inputsToBuild[key] = t
	} else if objType == OUTPUT_TYPE {
		s.outputsToBuild[key] = t
	} else {
		panic(fmt.Sprintf("Invalid object type %s", objType))
	}
	s.findDependentObjects(t, objType)
}

func (s *SchemaBuilder) CreateObjectsFields() (map[string]graphql.Input, map[string]graphql.Output) {
	for n, t := range s.inputsToBuild {
		bo := s.builtInputs[n].(*graphql.InputObject)
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fName := getFieldName(f.Name)
			of := s.createInputField(f.Name, f.Type, true)
			bo.AddFieldConfig(fName, of)
		}
	}
	for n, t := range s.outputsToBuild {
		bo := s.builtOutputs[n].(*graphql.Object)
		co := s.customObjects[n]
		for i := 0; i < t.NumField(); i++ {
			var of *graphql.Field
			f := t.Field(i)
			fName := getFieldName(f.Name)
			if co != nil {
				_co := co.(*Object)
				if v, ok := _co.Methods[fName]; ok {
					of = s.buildMethod(fName, v, _co)
				} else {
					of = s.createOutputField(f.Name, f.Type, true)
				}
			} else {
				of = s.createOutputField(f.Name, f.Type, true)
			}
			bo.AddFieldConfig(fName, of)
		}
	}
	return s.builtInputs, s.builtOutputs
}

func (s *SchemaBuilder) createInputField(fieldName string, t reflect.Type, required bool) *graphql.InputObjectFieldConfig {
	fType := s.getInputFieldTypeRecursive(t, true)
	if fType == nil {
		log.Errorf("Cannot create input field %s", fieldName)
		return nil
	}
	field := &graphql.InputObjectFieldConfig{
		Type: fType,
	}
	return field
}

func (s *SchemaBuilder) createOutputField(fieldName string, t reflect.Type, required bool) *graphql.Field {
	fType := s.getOutputFieldTypeRecursive(t, true)
	if fType == nil {
		log.Errorf("Cannot create output field %s", fieldName)
		return nil
	}
	field := &graphql.Field{
		Name: fieldName,
		Type: fType,
	}
	return field
}

func (s *SchemaBuilder) getOutputFieldType(v graphql.Output, required bool) graphql.Output {
	if required {
		return graphql.NewNonNull(v)
	}

	return v

}

func (s *SchemaBuilder) getInputFieldType(v graphql.Input, required bool) graphql.Input {
	if required {
		return graphql.NewNonNull(v)
	}

	return v

}

func (s *SchemaBuilder) getInputFieldTypeRecursive(t reflect.Type, required bool) graphql.Input {
	switch t.Kind() {

	case reflect.Ptr:
		return s.getInputFieldTypeRecursive(t.Elem(), false)
	case reflect.Slice:
		if v, ok := s.isScalar(t); ok {
			return s.getInputFieldType(v, required)
		} else {
			return graphql.NewList(s.getInputFieldTypeRecursive(t.Elem(), true))
		}
	case reflect.Struct:
		if v, ok := s.isScalar(t); ok {
			return s.getInputFieldType(v, required)
		} else {
			key := getKey(t)
			v := s.builtInputs[key]
			return s.getInputFieldType(v, required)
		}
	}

	if v, ok := s.isScalar(t); ok {
		return s.getInputFieldType(v, required)
	}

	return nil
}

func (s *SchemaBuilder) getOutputFieldTypeRecursive(t reflect.Type, required bool) graphql.Output {
	//var field *graphql.Field
	switch t.Kind() {

	case reflect.Ptr:
		return s.getOutputFieldTypeRecursive(t.Elem(), false)
	case reflect.Slice:
		if v, ok := s.isScalar(t); ok {
			return s.getOutputFieldType(v, required)
		} else {
			return graphql.NewList(s.getOutputFieldTypeRecursive(t.Elem(), true))
		}
	case reflect.Struct:
		if v, ok := s.isScalar(t); ok {
			return s.getOutputFieldType(v, required)
		} else {
			key := getKey(t)
			v := s.builtOutputs[key]
			return s.getOutputFieldType(v, required)
		}
	}

	if v, ok := s.isScalar(t); ok {
		return s.getOutputFieldType(v, required)
	}

	return nil
}

func (s *SchemaBuilder) buildMethod(n string, v *Method, o *Object) *graphql.Field {
	out := s.getResolverOutputObject(v.Fn)
	args := s.getResolverArgs(v.Fn)
	var fieldConfigArgument graphql.FieldConfigArgument

	if args != nil {
		if s.argsMap == nil {
			s.argsMap = make(map[string]map[string]interface{})
		}

		if s.argsMap[o.Name] == nil {
			s.argsMap[o.Name] = make(map[string]interface{})
		}
		s.argsMap[o.Name][n] = reflect.New(args).Elem().Interface()
		fieldConfigArgument = s.buildFieldConfigArgument(args)
	}

	fun := s.getFunc(v.Fn)
	return &graphql.Field{
		Args: fieldConfigArgument,
		Type: out,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if p.Context == nil {
				p.Context = context.Background()
			}
			if _, ok := p.Source.(map[string]interface{}); ok {
				selection := ParseSelections(p, s.argsMap)
				p.Context = context.WithValue(p.Context, "selection", selection)
			}
			in := make([]reflect.Value, fun.Type().NumIn())

			in[0] = reflect.ValueOf(p.Context)

			argType, pos, _ := getArgs(fun.Type())

			if _, ok := p.Source.(map[string]interface{}); !ok {
				in[pos-1] = reflect.ValueOf(p.Source)
			}

			if p.Args != nil && len(p.Args) > 0 {
				args := ReflectStructRecursive(argType, p.Args)
				in[pos] = args
			} else {
				if argType != nil {
					in[pos] = reflect.New(argType).Elem()
				}
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

func (s *SchemaBuilder) buildMethods(o *Object) graphql.Fields {
	if s.argsMap == nil {
		s.argsMap = make(map[string]map[string]interface{})
	}
	fields := graphql.Fields{}
	for n, v := range o.Methods {
		fields[n] = s.buildMethod(n, v, o)
	}
	return fields
}

func (s *SchemaBuilder) buildSubscriptionMethods(so *SubscriptionObject) graphql.Fields {
	fields := graphql.Fields{}
	for n, v := range so.Methods {
		out, _ := s.getResolverOutputObjectFromType(reflect.TypeOf(v.Output))
		args := s.getResolverArgs(v.Fn)

		var fieldConfigArgument graphql.FieldConfigArgument

		if args != nil {
			if s.argsMap == nil {
				s.argsMap = make(map[string]map[string]interface{})
			}

			if s.argsMap[v.Name] == nil {
				s.argsMap[v.Name] = make(map[string]interface{})
			}
			s.argsMap[v.Name][n] = reflect.New(args).Elem().Interface()
			fieldConfigArgument = s.buildFieldConfigArgument(args)
		}

		fun := s.getFunc(v.Fn)
		fields[n] = &graphql.Field{
			Args: fieldConfigArgument,
			Type: out,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return p.Source, nil
			},
			Subscribe: func(p graphql.ResolveParams) (interface{}, error) {
				c := make(chan interface{})
				in := make([]reflect.Value, fun.Type().NumIn())
				if p.Context == nil {
					in[0] = reflect.New(fun.Type().In(0)).Elem()
				} else {
					in[0] = reflect.ValueOf(p.Context)
				}
				in[1] = reflect.ValueOf(c)
				if p.Args != nil && len(p.Args) > 0 {
					argType, _, _ := getArgs(fun.Type())
					args := ReflectStructRecursive(argType, p.Args)
					in[2] = args
				}

				go func() {
					fun.Call(in)
				}()
				return c, nil
			},
		}
	}
	return fields
}

func (s *SchemaBuilder) findResolverOutputObject(fn interface{}) reflect.Type {
	rf := reflect.TypeOf(fn).Out(0)
	return s.getActualTypeRecursive(rf)
}

func (s *SchemaBuilder) findSubscriptionOutputObject(out interface{}) reflect.Type {
	rf := reflect.TypeOf(out)
	return s.getActualTypeRecursive(rf)
}

func (s *SchemaBuilder) findDependentObjects(t reflect.Type, objType string) {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		log.Println(f.Name)
		ao := s.getActualTypeRecursive(f.Type)

		log.Println(f.Type)
		_, scalar := s.isScalar(ao)
		if !scalar {
			key := getKey(ao)
			if objType == INPUT_TYPE {
				if _, ok := s.inputsToBuild[key]; ok {
					continue
				}
				s.inputsToBuild[key] = ao
			} else if objType == OUTPUT_TYPE {
				if _, ok := s.outputsToBuild[key]; ok {
					continue
				}
				s.outputsToBuild[key] = ao
			} else {
				panic(fmt.Sprintf("Invalid object type %s", objType))
			}

			s.findDependentObjects(ao, objType)

			//if _, ok := s.outputsToBuild[key]; !ok {
			//	if objType == INPUT_TYPE {
			//		s.inputsToBuild[key] = ao
			//	} else if objType == OUTPUT_TYPE {
			//		s.outputsToBuild[key] = ao
			//	} else {
			//		panic(fmt.Sprintf("Invalid object type %s", objType))
			//	}
			//
			//	s.findDependentObjects(ao, objType)
			//}
		}
	}
}

func (s *SchemaBuilder) getActualTypeRecursive(t reflect.Type) reflect.Type {
	switch t.Kind() {
	case reflect.Ptr:
		return s.getActualTypeRecursive(t.Elem())
	case reflect.Slice:
		_, scalar := s.isScalar(t)
		if scalar {
			return t
		}
		return s.getActualTypeRecursive(t.Elem())

	case reflect.Struct:
		return t
	}
	return t

}

func (s *SchemaBuilder) findResolverArgsObject(fn interface{}) reflect.Type {
	args, _, exist := getArgs(reflect.TypeOf(fn))
	if !exist {
		return nil
	}

	return s.getActualTypeRecursive(args)
}

func (s *SchemaBuilder) getResolverArgs(fn interface{}) reflect.Type {
	a := s.findResolverArgsObject(fn)
	return a
}

func (s *SchemaBuilder) getResolverOutputObject(fn interface{}) graphql.Output {
	rf := reflect.TypeOf(fn).Out(0)
	return s.getResolverOutputObjectRecursive(rf)
}

func (s *SchemaBuilder) getResolverOutputObjectRecursive(t reflect.Type) graphql.Output {
	switch t.Kind() {
	case reflect.Ptr:
		return MakeObjectNullable(s.getResolverOutputObjectRecursive(t.Elem()))
	case reflect.Slice:
		return graphql.NewNonNull(graphql.NewList(s.getResolverOutputObjectRecursive(t.Elem())))
	case reflect.Struct:
		return graphql.NewNonNull(s.builtOutputs[getKey(t)])
	}

	panic("Invalid output type")
}

func (s *SchemaBuilder) getResolverInputObjectRecursive(t reflect.Type) graphql.Input {
	switch t.Kind() {
	case reflect.Ptr:
		return MakeObjectNullable(s.getResolverInputObjectRecursive(t.Elem()))
	case reflect.Slice:
		return graphql.NewNonNull(graphql.NewList(s.getResolverInputObjectRecursive(t.Elem())))
	case reflect.Struct:
		return graphql.NewNonNull(s.builtInputs[getKey(t)])
	}
	if sc, ok := s.isScalar(t); ok {
		return graphql.NewNonNull(sc)
	}

	panic("Invalid input type")
}

func (s *SchemaBuilder) getResolverOutputObjectFromType(t reflect.Type) (graphql.Output, reflect.Type) {

	return s.builtOutputs[getKey(t)], t
}

func (s *SchemaBuilder) RegisterScalar(key string, sType *graphql.Scalar) {
	s.checkScalars(key)
	s.scalars[key] = sType
}

func (s *SchemaBuilder) SetDefaultScalars() {
	if s.scalars == nil {
		s.scalars = make(map[string]*graphql.Scalar)
	}

	for k, v := range defaultScalarsMap {
		if _, ok := s.scalars[k]; !ok {
			s.scalars[k] = v
		}
	}
}

func (s *SchemaBuilder) isScalar(t reflect.Type) (*graphql.Scalar, bool) {
	n := t.Name()
	if v, ok := s.scalars[n]; ok {
		return v, true
	}
	return nil, false
}

func (s *SchemaBuilder) getFunc(fn interface{}) reflect.Value {
	rf := reflect.ValueOf(fn)
	return rf
}

func (s *SchemaBuilder) Build() (graphql.Schema, error) {
	s.SetDefaultScalars()
	s.FindObjectsToBuild()
	s.CreateObjects()
	s.CreateObjectsFields()

	mutation := s.buildMutation()
	query := s.buildQuery()
	subscription := s.buildSubscription()

	schemaConfig := graphql.SchemaConfig{Query: query, Mutation: mutation, Subscription: subscription}
	schema, err := graphql.NewSchema(schemaConfig)

	if err != nil {
		logger.GetLogger().Error("Error build Gomer schema", err)
		return schema, err
	}
	logger.GetLogger().Infoln("Gomer schema build successfully")

	return schema, err
}

type FieldResolveFn func() (interface{}, error)
