package test_uttils

import (
	"github.com/shopspring/decimal"
	"time"
)

type Ticket struct {
	ID         string          `json:"id"`
	Title      string          `json:"title"`
	Number     int             `json:"number"`
	Time       time.Time       `json:"time"`
	Decimal    decimal.Decimal `json:"decimal"`
	Categories []*Category     `json:"categories"`
	Tags       []*Tag          `json:"tags"`
}

type StringFilter struct {
	Neq  *string `json:"neq"`
	Eq   *string `json:"eq"`
	Like *string `json:"like"`
}

type NumberFilter struct {
	Neq *int `json:"neq"`
	Eq  *int `json:"eq"`
	Lt  *int `json:"lt"`
	Gt  *int `json:"gt"`
	Gte *int `json:"gte"`
	Lte *int `json:"lte"`
}

type IDFilter struct {
	Neq *int64   `json:"neq"`
	Eq  *int64   `json:"eq"`
	In  *[]int64 `json:"in"`
}

type TicketFilterInput struct {
	ID     *IDFilter             `json:"id"`
	Title  *StringFilter         `json:"title"`
	Number *NumberFilter         `json:"number"`
	And    *[]*TicketFilterInput `gomer:"ignoreFields:And,Or;prefix:Inner"`
}

type TicketOrderInput struct {
	Title  *string
	Number string
}

type TagFilterInput struct {
	Title  *StringFilter
	Number *NumberFilter
}

type TagOrderInput struct {
	Title  *string
	Number string
}

type ItemInterface struct {
	Title   string
	Number  *int
	Numbers []*int64
}

type TicketInsertInput struct {
	Title           *string
	Number          *int
	Numbers         []*int64
	NumbersRequired []int64
	Decimal         *decimal.Decimal
}

type TicketUpdateInput struct {
	ID     string
	Title  *string
	Number *int
}

type Status struct {
	ID    string
	Title string
}

type Category struct {
	ID    string
	Title string
}

type Tag struct {
	ID    string
	Title string
}
