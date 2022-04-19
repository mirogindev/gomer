package gqbuilder

import (
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
)

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

type GomerTags map[string]string

//func (gt GomerTags) FromType(t reflect.StructField) {
//
//	}
//}

func (gt GomerTags) GetParam(n string) string {
	if v, ok := gt[n]; ok {
		return v
	}

	log.Errorf("Param %s does not exist", n)
	return ""

}

func (gt GomerTags) ParamExist(n string) (string, bool) {
	if v, ok := gt[n]; ok {
		return v, true
	}
	return "", false
}
