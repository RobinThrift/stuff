package views

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"golang.org/x/net/html"
)

type componentFS struct {
	fs fs.FS
}

func (cfs *componentFS) Open(name string) (fs.File, error) {
	file, err := cfs.fs.Open(name)
	if err != nil {
		return nil, err
	}

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return file, nil
	}

	var converted bytes.Buffer
	r := newReplacer(file)

	err = r.replace(&converted)
	if err != nil {
		panic(fmt.Errorf("%s: %w", stat.Name(), err))
	}

	return &componentFile{File: file, converted: converted}, nil
}

type componentFile struct {
	fs.File
	converted bytes.Buffer
}

func (cf *componentFile) Read(p []byte) (int, error) {
	return cf.converted.Read(p)
}

type replacer struct {
	tokenizer *html.Tokenizer
	buf       bytes.Buffer
	raw       bytes.Buffer
	depth     int

	childTemplates map[int]*bytes.Buffer
	tagCounter     map[string]int
}

func newReplacer(r io.Reader) *replacer {
	return &replacer{
		tokenizer:      html.NewTokenizer(r),
		tagCounter:     map[string]int{},
		childTemplates: map[int]*bytes.Buffer{},
	}
}

func (r *replacer) replace(w io.Writer) error {
	for {
		r.buf.Reset()
		r.raw.Reset()

		tt := r.tokenizer.Next()
		_, err := r.raw.Write(r.tokenizer.Raw())
		if err != nil {
			return err
		}

		err = r.processToken(w, tt)
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}
	}

	for _, c := range r.childTemplates {
		_, err := c.WriteTo(w)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *replacer) processToken(w io.Writer, tt html.TokenType) error {
	switch tt { //nolint: exhaustive
	case html.ErrorToken:
		return r.tokenizer.Err()
	case html.SelfClosingTagToken:
		token := r.tokenizer.Token()
		if !isComponent(token.Data) {
			break
		}

		err := convertTokenToTemplateCall(token, false, 0, &r.buf)
		if err != nil {
			return err
		}

		_, err = r.buf.WriteTo(w)
		return err
	case html.StartTagToken:
		token := r.tokenizer.Token()
		if !isComponent(token.Data) {
			break
		}

		r.tagCounter[token.Data]++

		err := convertTokenToTemplateCall(token, true, r.tagCounter[token.Data], &r.buf)
		if err != nil {
			return err
		}

		childBuf, err := prepareChildTemplate(token.Data, r.tagCounter[token.Data])
		if err != nil {
			return err
		}
		r.depth++
		r.childTemplates[r.depth] = childBuf

		_, err = r.buf.WriteTo(w)
		return err
	case html.EndTagToken:
		token := r.tokenizer.Token()
		if !isComponent(token.Data) {
			break
		}

		buf := r.childTemplates[r.depth]
		err := endChildTemplate(buf)
		if err != nil {
			return err
		}
		_, err = r.buf.WriteTo(w)
		if err != nil {
			return err
		}

		r.depth--
		if r.depth == 0 {
			return nil
		}
	}

	if r.depth != 0 {
		buf := r.childTemplates[r.depth]
		_, err := r.raw.WriteTo(buf)
		return err
	}

	_, err := r.raw.WriteTo(w)
	return err
}

func convertTokenToTemplateCall(token html.Token, hasChildren bool, counter int, buf *bytes.Buffer) error {
	if len(token.Attr) == 0 {
		_, err := fmt.Fprintf(buf, `{{ template "%s" }}`, token.Data[2:])
		return err
	}

	_, err := fmt.Fprintf(buf, `{{ template "%s" dict `, token.Data[2:])
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(buf, `"%s" `, token.Attr[0].Key)
	if err != nil {
		return err
	}

	err = valueToTemplate(token.Attr[0].Val, buf)
	if err != nil {
		return err
	}

	for _, attr := range token.Attr[1:] {
		if len(attr.Val) == 0 {
			continue
		}

		_, err = fmt.Fprintf(buf, ` "%s" `, attr.Key)
		if err != nil {
			return err
		}

		switch {
		case len(attr.Val) > 1 && attr.Val[0] == '{' && attr.Val[1] != '{':
			err = jsonMapToTemplate(attr.Val, buf)
		case attr.Val[0] == '[':
			err = jsonArrayToTemplate(attr.Val, buf)
		default:
			err = valueToTemplate(attr.Val, buf)
		}

		if err != nil {
			return err
		}
	}

	if hasChildren {
		_, err = fmt.Fprintf(buf, ` "children" (children "%s-children-%d" .)`, token.Data, counter)
		if err != nil {
			return err
		}
	}

	_, err = buf.WriteString(" }}")
	return err
}

func prepareChildTemplate(tag string, counter int) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	_, err := fmt.Fprintf(&buf, "\n{{ define \"%s-children-%d\" }}", tag, counter)
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func endChildTemplate(buf *bytes.Buffer) error {
	_, err := buf.WriteString("{{ end }}\n")
	return err
}

func jsonMapToTemplate(val string, b *bytes.Buffer) error {
	m := map[string]any{}

	err := json.Unmarshal([]byte(val), &m)
	if err != nil {
		return err
	}

	return valueMapToTemplate(m, b)
}

func jsonArrayToTemplate(val string, b *bytes.Buffer) error {
	v := []any{}

	err := json.Unmarshal([]byte(val), &v)
	if err != nil {
		return err
	}

	return valueSliceToTemplate(v, b)
}

func valueToTemplate(v any, b *bytes.Buffer) error {
	switch v := v.(type) {
	case []any:
		return valueSliceToTemplate(v, b)
	case map[string]any:
		return valueMapToTemplate(v, b)
	case string:
		if len(v) > 1 && v[0] == '{' && v[1] == '{' {
			_, err := fmt.Fprintf(b, `(%s)`, v[2:len(v)-2])
			return err
		}

		if len(v) == 0 || v[0] != '(' {
			_, err := fmt.Fprintf(b, `"%s"`, v)
			return err
		}
	}

	_, err := fmt.Fprint(b, v)
	return err
}

func valueSliceToTemplate(v []any, b *bytes.Buffer) error {
	if len(v) == 0 {
		_, err := b.WriteString("(list)")
		return err
	}

	_, err := b.WriteString("(list ")
	if err != nil {
		return err
	}

	err = valueToTemplate(v[0], b)
	if err != nil {
		return err
	}

	for _, item := range v[1:] {
		_, err = b.WriteRune(' ')
		if err != nil {
			return err
		}

		err = valueToTemplate(item, b)
		if err != nil {
			return err
		}
	}

	_, err = b.WriteRune(')')
	return err
}

func valueMapToTemplate(m map[string]any, b *bytes.Buffer) error {
	if len(m) == 0 {
		_, err := b.WriteString("(dict)")
		return err
	}

	_, err := b.WriteString("(dict")
	if err != nil {
		return err
	}

	for key, value := range m {
		_, err = b.WriteRune(' ')
		if err != nil {
			return err
		}

		_, err = fmt.Fprintf(b, `"%s" `, key)
		if err != nil {
			return err
		}

		err = valueToTemplate(value, b)
		if err != nil {
			return err
		}
	}

	_, err = b.WriteRune(')')
	return err
}

func isComponent(tag string) bool {
	return strings.HasPrefix(tag, "x-")
}
