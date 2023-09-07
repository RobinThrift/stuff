package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeSearchQuery(t *testing.T) {
	tt := []struct {
		in  string
		exp *ListAssetsQuerySearch
	}{
		{in: "", exp: &ListAssetsQuerySearch{Fields: map[string]string{}}},
		{in: "test", exp: &ListAssetsQuerySearch{Raw: "test", Fields: map[string]string{}}},
		{in: "test with spaces", exp: &ListAssetsQuerySearch{Raw: "test with spaces", Fields: map[string]string{}}},
		{in: "name: name goes here", exp: &ListAssetsQuerySearch{
			Raw: "name: name goes here",
			Fields: map[string]string{
				"name": "name goes here",
			},
		}},
		{
			in: "name: name goes here category: cool stuff",
			exp: &ListAssetsQuerySearch{
				Raw: "name: name goes here category: cool stuff",
				Fields: map[string]string{
					"name":     "name goes here",
					"category": "cool stuff",
				},
			},
		},
		{
			in: "name:nospace category: space here",
			exp: &ListAssetsQuerySearch{
				Raw: "name:nospace category: space here",
				Fields: map[string]string{
					"name":     "nospace",
					"category": "space here",
				},
			},
		},
		{
			in: "tag: taghere name: namehere category: categroyhere model: modelhere ModelNo: modelnohere serialno: serialnohere manufacturer: manufacturerhere",
			exp: &ListAssetsQuerySearch{
				Raw: "tag: taghere name: namehere category: categroyhere model: modelhere ModelNo: modelnohere serialno: serialnohere manufacturer: manufacturerhere",
				Fields: map[string]string{
					"tag":          "taghere",
					"name":         "namehere",
					"category":     "categroyhere",
					"model":        "modelhere",
					"modelno":      "modelnohere",
					"serialno":     "serialnohere",
					"manufacturer": "manufacturerhere",
				},
			},
		},
	}

	for _, tt := range tt {
		assert.Equal(t, tt.exp, decodeSearchQuery(tt.in))
	}
}
