package gqbuilder

import "github.com/graphql-go/graphql"

type GomerObject interface {
	GetType() interface{}
}

type SelectionSet struct {
	Selections []*Selection
}

type Selection struct {
	Name         string
	Alias        string
	Args         interface{}
	SelectionSet *SelectionSet
}

type GomerInputObject struct {
	graphql.InputObject
	ArgsObject interface{}
}
