package models

type Ticket struct {
	ID         string
	Title      string
	Number     int
	Categories []*Category
	Tag        *Tag
}

type TicketFilterInput struct {
	Title  string
	Number *int
}

type TicketOrderInput struct {
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
