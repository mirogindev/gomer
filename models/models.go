package models

import (
	"github.com/shopspring/decimal"
	"time"
)

type Ticket struct {
	ID         string
	Title      string
	Number     int
	Time       time.Time
	Decimal    decimal.Decimal
	Categories []*Category
	Tags       []*Tag
}

type StringFilter struct {
	Neq  *string `json:"neq" mapstructure:"eq" op:"<>"`
	Eq   *string `json:"eq" mapstructure:"eq" op:"="`
	Like *string `json:"like" mapstructure:"like" op:"like"`
}

type NumberFilter struct {
	Neq *int `json:"neq"`
	Eq  *int `json:"eq"`
	Lt  *int `json:"lt"`
	Gt  *int `json:"gt"`
	Gte *int `json:"gte"`
	Lte *int `json:"lte"`
}

type TicketFilterInput struct {
	Title  *StringFilter `json:"title"`
	Number *NumberFilter `json:"number"`
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

type TicketInsertInput struct {
	Title  string
	Number *int
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
