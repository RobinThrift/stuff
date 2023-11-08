package htmlui

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

func decodeParams(target any, r *http.Request) error {
	query := r.URL.Query()

	val := unpackTarget(target)

	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		if urlTag, ok := field.Tag.Lookup("url"); ok {
			if err := setField(val.Field(i), chi.URLParam(r, urlTag)); err != nil {
				return fmt.Errorf("error parsing url parameter %s as %v", urlTag, val.Field(i).Type())
			}
			continue
		}

		if queryTag, ok := field.Tag.Lookup("query"); ok {
			if err := setField(val.Field(i), query.Get(queryTag)); err != nil {
				return fmt.Errorf("error parsing query parameter %s as %v", query, val.Field(i).Type())
			}
			continue
		}
	}

	return nil
}

func setField(target reflect.Value, val string) error {
	if val == "" {
		return nil
	}

	switch target.Type().Kind() { //nolint: exhaustive
	case reflect.String:
		target.SetString(val)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		target.SetBool(b)
	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		target.SetInt(i)
	}

	return nil
}

func decodeSearchQuery(queryStr string) map[string]string {
	queryStr = strings.TrimPrefix(queryStr, "*")
	fields := map[string]string{}

	lastWordEnd := 0
	lastNameEnd := 0
	name := ""
	value := ""
	for i := 0; i < len(queryStr)-1; i++ {
		switch queryStr[i] {
		case ':':
			value = queryStr[lastNameEnd:lastWordEnd]
			if name != "" {
				fields[strings.ToLower(name)] = value
			}
			if queryStr[lastWordEnd] == ' ' {
				lastWordEnd += 1
			}
			name = queryStr[lastWordEnd:i]
			lastNameEnd = i + 1
			if i+1 < len(queryStr) && queryStr[i+1] == ' ' {
				lastNameEnd = i + 2
			}
		case ' ':
			lastWordEnd = i
		}
	}

	if name != "" {
		value = queryStr[lastNameEnd:]
		if value != "" {
			fields[strings.ToLower(name)] = queryStr[lastNameEnd:]
		}
	}

	return fields
}

func unpackTarget(target any) reflect.Value {
	val := reflect.ValueOf(target)

	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	return val
}
