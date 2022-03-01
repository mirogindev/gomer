package models

type Ticket struct {
	ID         string
	Title      string
	Number     *int
	Categories []*Category
	Tag        *Tag
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

type Category struct {
	ID    string
	Title string
}

type Tag struct {
	ID    string
	Title string
}
