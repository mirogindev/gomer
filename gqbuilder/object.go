package gqbuilder

import "fmt"

type Object struct {
	Name        string
	Type        interface{}
	Description string
	Resolver    interface{}
	Args        interface{}
	Methods     map[string]*Method
}

func (s *Object) FieldFunc(name string, handler interface{}) {
	s.checkMethods(name)

	s.Methods[name] = &Method{
		Name: name,
		Fn:   handler,
	}
}

func (s *Object) checkMethods(name string) {
	if s.Methods == nil {
		s.Methods = make(map[string]*Method)
	}
	if s.Methods[name] != nil {
		panic(fmt.Sprintf("Func with name %s aready exists", name))
	}
}

type Method struct {
	Name string
	Fn   interface{}
}
