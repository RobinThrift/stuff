package htmlui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecodeSearchQuery(t *testing.T) {
	tt := []struct {
		in  string
		exp map[string]string
	}{
		{in: "", exp: map[string]string{}},
		{in: "test", exp: map[string]string{}},
		{in: "test with spaces", exp: map[string]string{}},
		{in: "name: name goes here", exp: map[string]string{"name": "name goes here"}},
		{
			in: "name: name goes here category: cool stuff",
			exp: map[string]string{
				"name":     "name goes here",
				"category": "cool stuff",
			},
		},
		{
			in: "name:nospace category: space here",
			exp: map[string]string{
				"name":     "nospace",
				"category": "space here",
			},
		},
		{
			in: "tag: taghere name: namehere category: categroyhere model: modelhere ModelNo: modelnohere serialno: serialnohere manufacturer: manufacturerhere",
			exp: map[string]string{
				"tag":          "taghere",
				"name":         "namehere",
				"category":     "categroyhere",
				"model":        "modelhere",
				"modelno":      "modelnohere",
				"serialno":     "serialnohere",
				"manufacturer": "manufacturerhere",
			},
		},
	}

	for _, tt := range tt {
		assert.Equal(t, tt.exp, decodeSearchQuery(tt.in))
	}
}
