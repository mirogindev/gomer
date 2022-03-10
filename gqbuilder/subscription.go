package gqbuilder

import (
	log "github.com/sirupsen/logrus"
)

type SubscriptionObject struct {
	Name        string
	Description string
	Resolver    interface{}
	Args        interface{}
	Methods     map[string]*SubscriptionMethod
}

func (s *SubscriptionObject) FieldSubscription(name string, output interface{}, handler interface{}) {
	s.checkMethods(name)

	s.Methods[name] = &SubscriptionMethod{
		Name:   name,
		Output: output,
		Fn:     handler,
	}
}

func (s *SubscriptionObject) checkMethods(name string) {
	if s.Methods == nil {
		s.Methods = make(map[string]*SubscriptionMethod)
	}

	if _, ok := s.Methods[name]; ok {
		log.Panicf("Func with name %s aready exists", name)
	}
}

type SubscriptionMethod struct {
	Name   string
	Output interface{}
	Fn     interface{}
}
