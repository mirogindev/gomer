package models

type Ticket struct {
	ID         string
	Title      string
	Number     int
	Categories []*Category
	Tags       []*Tag
}

type StringFilter struct {
	Neq  *string `json:"neq" mapstructure:"eq" op:"<>"`
	Eq   *string `json:"eq" mapstructure:"eq" op:"="`
	Like *string `json:"like" mapstructure:"like" op:"like"`
}

type TicketFilterInput struct {
	Title  *StringFilter `json:"name" mapstructure:"name"`
	Number *int
}

type TicketOrderInput struct {
	Title  *string
	Number string
}

type TagFilterInput struct {
	Title  string
	Number *int
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
