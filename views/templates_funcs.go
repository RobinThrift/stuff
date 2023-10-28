package views

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

var templateFuncs = template.FuncMap{
	"list": func(items ...any) []any {
		return items
	},

	"split": strings.Split,

	"dict": func(pairs ...any) map[string]any {
		dict := map[string]any{}

		lenPairs := len(pairs)
		for i := 0; i < lenPairs; i += 2 {
			key := fmt.Sprint(pairs[i])

			if i+1 >= lenPairs {
				continue
			}

			dict[key] = pairs[i+1]
		}

		return dict
	},

	"merge": func(a, b map[string]any) map[string]any {
		for k, v := range b {
			a[k] = v
		}

		return a
	},

	"has": func(m any, k string) bool {
		switch m := m.(type) {
		case map[string]any:
			_, ok := m[k]
			return ok
		case map[string]string:
			_, ok := m[k]
			return ok
		}

		return false
	},

	"get": func(m any, k string) any {
		switch m := m.(type) {
		case map[string]any:
			return m[k]
		case map[string]string:
			return m[k]
		}

		return nil
	},

	"default": func(val any, d any) any {
		if isZeroValue(val) {
			return d
		}

		return val
	},

	"mul": func(a int, b int) int {
		return a * b
	},

	"add": func(a int, b int) int {
		return a + b
	},

	"min": func(a int, b int) int { //nolint: gocritic // false positive because new builtins
		return min(a, b)
	},

	"max": func(a int, b int) int { //nolint: gocritic // false positive because new builtins
		return max(a, b)
	},

	"between": func(a, lower, upper int) bool {
		return a > lower && a < upper
	},

	"json": func(v any) template.JS {
		j, err := json.Marshal(v)
		if err != nil {
			panic(err)
		}
		return template.JS(j)
	},

	"debug": func(v any) string {
		j, err := json.MarshalIndent(v, "", "    ")
		if err != nil {
			panic(err)
		}
		return string(j)
	},

	"bytesToPNG": func(b []byte) template.URL {
		return template.URL("data:image/png;base64," + base64.StdEncoding.EncodeToString(b))
	},

	"sequence": func(from int, to int) ([]int, error) {
		if from > to {
			return nil, errors.New("from is greater than to")
		}

		r := make([]int, to-from+1)

		i := 0
		for c := from; c <= to; c++ {
			r[i] = c
			i++
		}

		return r, nil
	},
}

// Adapted from https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/defaults.go#L35
// LICENSE MIT (https://github.com/Masterminds/sprig/blob/581758eb7d96ae4d113649668fa96acc74d46e7f/LICENSE.txt)
func isZeroValue(value any) bool {
	if s, ok := value.(string); ok {
		return s == ""
	}

	val := reflect.ValueOf(value)

	// Basically adapted from text/template.isTrue
	switch val.Kind() { //nolint: exhaustive
	case reflect.Invalid:
		return true
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return val.Len() == 0
	case reflect.Bool:
		return !val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return val.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return val.Float() == 0
	case reflect.Struct:
		return false
	case reflect.Pointer:
		return val.IsNil()
	default:
		panic(fmt.Sprintf("can't check default value for %T", value))
	}
}
