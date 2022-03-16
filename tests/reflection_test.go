package tests

import (
	"github.com/mirogindev/gomer/gqbuilder"
	"github.com/mirogindev/gomer/models"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type Args struct {
	Filter             *models.TicketFilterInput
	Order              models.TicketOrderInput
	TagsPointers       []*models.Tag
	Tags               []models.Tag
	PointerTags        *[]models.Tag
	PointerTagsPointer *[]*models.Tag
	Limit              int
	Offset             *int
}

func TestSimpleFieldsReflection(t *testing.T) {
	log.SetLevel(log.TraceLevel)
	params := make(map[string]interface{})
	offset := 10
	limit := 15

	params["limit"] = limit
	params["offset"] = offset

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Equal(t, obj.Limit, limit)
	assert.Equal(t, *obj.Offset, offset)

}

func TestFieldReflectionWithNestedStruct(t *testing.T) {
	params := make(map[string]interface{})
	order := make(map[string]interface{})
	title := "desc"
	number := "asc"

	order["title"] = title
	order["number"] = number

	params["order"] = order

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Equal(t, *obj.Order.Title, title)
	assert.Equal(t, obj.Order.Number, number)

}

func TestFieldReflectionWithNestedPointerStruct(t *testing.T) {
	params := make(map[string]interface{})
	filter := make(map[string]interface{})
	titleFilter := make(map[string]interface{})
	numberFilter := make(map[string]interface{})
	title := "testTitle"
	number := 20

	titleFilter["eq"] = title
	numberFilter["eq"] = number

	filter["title"] = titleFilter
	filter["number"] = numberFilter

	params["filter"] = filter

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.NotEmpty(t, obj.Filter)

	assert.Equal(t, *obj.Filter.Title.Eq, title)
	assert.Equal(t, *obj.Filter.Number.Eq, number)

}

func TestFieldReflectionWithNestedListWithPointerStructs(t *testing.T) {
	params := make(map[string]interface{})
	tags := make([]interface{}, 0)
	title1 := "testTitle1"
	title2 := "testTitle2"
	title3 := "testTitle3"

	tag1 := make(map[string]interface{})
	tag2 := make(map[string]interface{})
	tag3 := make(map[string]interface{})

	tag1["title"] = title1
	tag2["title"] = title2
	tag3["title"] = title3

	tags = append(tags, tag1)
	tags = append(tags, tag2)
	tags = append(tags, tag3)

	params["tags_pointers"] = tags

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Len(t, obj.TagsPointers, 3)
	//assert.Equal(t, obj.Filter.Number, number)

}

func TestFieldReflectionWithNestedListWithStructs(t *testing.T) {
	params := make(map[string]interface{})
	tags := make([]interface{}, 0)
	title1 := "testTitle1"
	title2 := "testTitle2"
	title3 := "testTitle3"

	tag1 := make(map[string]interface{})
	tag2 := make(map[string]interface{})
	tag3 := make(map[string]interface{})

	tag1["title"] = title1
	tag2["title"] = title2
	tag3["title"] = title3

	tags = append(tags, tag1)
	tags = append(tags, tag2)
	tags = append(tags, tag3)

	params["tags"] = tags

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Len(t, obj.Tags, 3)
}

func TestFieldReflectionWithNestedPointerListWithStructs(t *testing.T) {
	params := make(map[string]interface{})
	tags := make([]interface{}, 0)
	title1 := "testTitle1"
	title2 := "testTitle2"
	title3 := "testTitle3"

	tag1 := make(map[string]interface{})
	tag2 := make(map[string]interface{})
	tag3 := make(map[string]interface{})

	tag1["title"] = title1
	tag2["title"] = title2
	tag3["title"] = title3

	tags = append(tags, tag1)
	tags = append(tags, tag2)
	tags = append(tags, tag3)

	params["pointer_tags"] = tags

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Len(t, *obj.PointerTags, 3)
}

func TestFieldReflectionWithNestedPointerListWithPointerStructs(t *testing.T) {
	log.SetLevel(log.TraceLevel)
	params := make(map[string]interface{})
	tags := make([]interface{}, 0)
	title1 := "testTitle1"
	title2 := "testTitle2"
	title3 := "testTitle3"

	tag1 := make(map[string]interface{})
	tag2 := make(map[string]interface{})
	tag3 := make(map[string]interface{})

	tag1["title"] = title1
	tag2["title"] = title2
	tag3["title"] = title3

	tags = append(tags, tag1)
	tags = append(tags, tag2)
	tags = append(tags, tag3)

	params["pointer_tags_pointer"] = tags

	st := reflect.TypeOf(Args{})

	args := gqbuilder.ReflectStructRecursive(st, params)

	obj := args.Interface().(Args)

	assert.Len(t, *obj.PointerTagsPointer, 3)
	tagsPointers := *obj.PointerTagsPointer
	t1 := tagsPointers[0].Title
	t2 := tagsPointers[1].Title
	t3 := tagsPointers[2].Title
	assert.Equal(t, t1, title1)
	assert.Equal(t, t2, title2)
	assert.Equal(t, t3, title3)
}

//func TestFieldReflectionWithNestedPointerStruct(t *testing.T) {
//	params := make(map[string]interface{})
//	tags := make([]interface{},3)
//	title1 := "testTitle1"
//	title2 := "testTitle2"
//	title3 := "testTitle3"
//
//	tag1 := make(map[string]interface{})
//	tag2 := make(map[string]interface{})
//	tag3 := make(map[string]interface{})
//
//	tag1["title"] = title1
//	tag2["title"] = title2
//	tag3["title"] = title3
//
//	params["tags"] = tags
//
//	st := reflect.TypeOf(Args{})
//
//	args := gqbuilder.ReflectStruct(st, params)
//
//	obj := args.Interface().(Args)
//
//	assert.Len(t, obj.Limit, 3)
//	//assert.Equal(t, obj.Filter.Number, number)
//
//}
