package views

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceSelfClosing(t *testing.T) {
	input := `{{ template "layout.html.tmpl" . }}

{{ define "main" }}
<header class="my-5 flex items-end">
	<h1 class="font-extrabold md:text-2xl lg:text-4xl">Assets</h1>

	<div
		{{ if ne .ID "" -}}
		id="{{ .ID }}"
		{{- end -}}
		class="flex-1 flex flex-col md:flex-row justify-end items-center"
	>
		<a href="/assets/new" class="btn btn-primary inline-block">New Asset</a>

		<x-dropdown class="md:ms-2"
			button-text="Export"
			items='[
				{ "url": "/assets/export/csv"}
			]'
		/>
	</div>
</header>`

	expected := `{{ template "layout.html.tmpl" . }}

{{ define "main" }}
<header class="my-5 flex items-end">
	<h1 class="font-extrabold md:text-2xl lg:text-4xl">Assets</h1>

	<div
		{{ if ne .ID "" -}}
		id="{{ .ID }}"
		{{- end -}}
		class="flex-1 flex flex-col md:flex-row justify-end items-center"
	>
		<a href="/assets/new" class="btn btn-primary inline-block">New Asset</a>

		{{ template "dropdown" dict "class" "md:ms-2" "button-text" "Export" "items" (list (dict "url" "/assets/export/csv")) }}
	</div>
</header>`

	var actual bytes.Buffer

	r := newReplacer(strings.NewReader(input))

	err := r.replace(&actual)
	assert.NoError(t, err)

	assert.Equal(t, expected, actual.String())
}

func TestReplaceWithChildren(t *testing.T) {
	input := `
<div
	{{ if ne .ID "" -}}
	id="{{ .ID }}"
	{{- end -}}
	class="flex-1 flex flex-col md:flex-row justify-end items-center"
>
	<a href="/assets/new" class="btn btn-primary inline-block">New Asset</a>

	<x-dropdown class="md:ms-2" button-text="Export">
		<li><a href="/assets/export/csv">Export All (CSV)</a></li>
		<li><a href="/assets/export/json">Export All (JSON)</a></li>
	</x-dropdown>
</div>`

	expected := `
<div
	{{ if ne .ID "" -}}
	id="{{ .ID }}"
	{{- end -}}
	class="flex-1 flex flex-col md:flex-row justify-end items-center"
>
	<a href="/assets/new" class="btn btn-primary inline-block">New Asset</a>

	{{ template "dropdown" dict "class" "md:ms-2" "button-text" "Export" "children" (template "x-dropdown-children-1" .) }}
</div>
{{ define "x-dropdown-children-1" }}
		<li><a href="/assets/export/csv">Export All (CSV)</a></li>
		<li><a href="/assets/export/json">Export All (JSON)</a></li>
	{{ end }}
`

	var output bytes.Buffer

	r := newReplacer(strings.NewReader(input))

	err := r.replace(&output)
	assert.NoError(t, err)

	actual := output.String()

	assert.Equal(t, expected, actual)
}
