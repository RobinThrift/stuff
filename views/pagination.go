package views

import (
	"fmt"
	"net/url"

	"github.com/RobinThrift/stuff/entities"
)

type Pagination[Item any] struct {
	*entities.ListPage[Item]

	URL *url.URL
}

func (p *Pagination[Item]) PrevPageURL() string {
	if p.Page == 0 {
		return ""
	}

	return p.PageURL(p.Page - 1)
}

func (p *Pagination[Item]) NextPageURL() string {
	if p.Page >= p.NumPages-1 {
		return ""
	}

	return p.PageURL(p.Page + 1)
}

func (p *Pagination[Item]) PageURL(n int) string {
	query := p.URL.Query()
	query.Set("page", fmt.Sprint(n))
	query.Set("page_size", fmt.Sprint(p.PageSize))

	return p.URL.Path + "?" + query.Encode()
}

const ellipsis = "..."

func (p *Pagination[Item]) PaginationURLs() []PaginationURL {
	query := p.URL.Query()
	query.Set("page_size", fmt.Sprint(p.PageSize))

	urls := make([]PaginationURL, 0, 6)

	startEllipsis := false
	endEllipsis := false

	for i := 0; i < p.NumPages; i++ {
		if between(i, p.Page-3, p.Page+3) || between(i, p.NumPages-2, p.NumPages+3) || between(i, -1, 2) {
			query.Set("page", fmt.Sprint(i))

			urls = append(urls, PaginationURL{
				URL:       p.URL.Path + "?" + query.Encode(),
				Text:      fmt.Sprint(i + 1),
				IsCurrent: i == p.Page,
			})
			continue
		}

		if i < p.Page && !startEllipsis {
			urls = append(urls, PaginationURL{Text: ellipsis})
			startEllipsis = true
			continue
		}

		if i > p.Page && !endEllipsis {
			urls = append(urls, PaginationURL{Text: ellipsis})
			endEllipsis = true
			continue
		}
	}

	return urls
}

type PaginationURL struct {
	URL       string
	Text      string
	IsCurrent bool
}

func between(a, lower, upper int) bool {
	return a > lower && a < upper
}
