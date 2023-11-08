package entities

type ListPage[Item any] struct {
	Items    []Item
	Total    int
	NumPages int
	Page     int
	PageSize int
}
