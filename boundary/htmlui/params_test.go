package htmlui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestDecodeParams(t *testing.T) {
	type empty struct{}
	type urlAndQueryParams struct {
		TagOrID   string `url:"id"`
		FileID    int64  `url:"fileID"`
		Query     string `query:"query"`
		Page      int    `query:"page"`
		PageSize  int    `query:"page_size"`
		OrderBy   string `query:"order_by"`
		OrderDir  string `query:"order_dir"`
		AssetType string `query:"type"`
	}

	tt := []struct {
		name   string
		route  string
		input  string
		target any
		exp    any
	}{
		{
			name:  "all empty",
			route: "/test", input: "/test",
			target: &empty{},
			exp:    &empty{},
		},
		{
			name:  "with url parameter",
			route: "/assets/{id}", input: "/assets/abc",
			target: &assetsGetParams{},
			exp:    &assetsGetParams{TagOrID: "abc"},
		},
		{
			name:  "with multiple url parameters",
			route: "/assets/{id}/files/{fileID}", input: "/assets/abc/files/1234",
			target: &fileDeleteParams{},
			exp:    &fileDeleteParams{TagOrID: "abc", FileID: 1234},
		},
		{
			name:  "query parameters",
			route: "/assets", input: "/assets?query=test&page=1&page_size=25&order_by=name&order_dir=desc&type=component",
			target: &assetsListParams{},
			exp: &assetsListParams{
				Query:     "test",
				Page:      1,
				PageSize:  25,
				OrderBy:   "name",
				OrderDir:  "desc",
				AssetType: "component",
			},
		},
		{
			name:  "query parameters some missing",
			route: "/assets", input: "/assets?query=test&page=1&order_dir=desc&type=component",
			target: &assetsListParams{},
			exp: &assetsListParams{
				Query:     "test",
				Page:      1,
				OrderDir:  "desc",
				AssetType: "component",
			},
		},
		{
			name:  "url param and query parameters",
			route: "/assets/{id}/files/{fileID}", input: "/assets/abc/files/45678?query=test&page=1&page_size=25&order_by=name&order_dir=desc&type=component",
			target: &urlAndQueryParams{},
			exp: &urlAndQueryParams{
				TagOrID:   "abc",
				FileID:    45678,
				Query:     "test",
				Page:      1,
				PageSize:  25,
				OrderBy:   "name",
				OrderDir:  "desc",
				AssetType: "component",
			},
		},
	}

	for _, tt := range tt {
		t.Run(tt.name, func(t *testing.T) {
			router := chi.NewRouter()
			router.Get(tt.route, func(_ http.ResponseWriter, r *http.Request) {
				err := decodeParams(&tt.target, r)
				assert.NoError(t, err)
			})

			r, err := http.NewRequest(http.MethodGet, tt.input, nil)
			assert.NoError(t, err)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, r)

			assert.Equal(t, tt.exp, tt.target)
		})
	}
}
