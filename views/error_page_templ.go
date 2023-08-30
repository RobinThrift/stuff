// Code generated by templ@v0.2.316 DO NOT EDIT.

package views

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

import "fmt"

func ErrorPage(code int, title string, message string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div class=\"grid h-screen px-4 bg-white place-content-center\"><div class=\"text-center\"><h1 class=\"font-black text-9xl text-transparent rainbow bg-clip-text bg-gradient-to-bl\">")
		if err != nil {
			return err
		}
		var var_2 string = fmt.Sprint(code)
		_, err = templBuffer.WriteString(templ.EscapeString(var_2))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h1><p class=\"text-2xl font-bold tracking-tight text-gray-900 sm:text-4xl\">")
		if err != nil {
			return err
		}
		var var_3 string = title
		_, err = templBuffer.WriteString(templ.EscapeString(var_3))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</p><p class=\"mt-4 text-gray-500\">")
		if err != nil {
			return err
		}
		var var_4 string = message
		_, err = templBuffer.WriteString(templ.EscapeString(var_4))
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</p>")
		if err != nil {
			return err
		}
		var_5 := templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
			templBuffer, templIsBuffer := w.(*bytes.Buffer)
			if !templIsBuffer {
				templBuffer = templ.GetBuffer()
				defer templ.ReleaseBuffer(templBuffer)
			}
			var_6 := `Back Home`
			_, err = templBuffer.WriteString(var_6)
			if err != nil {
				return err
			}
			if !templIsBuffer {
				_, err = io.Copy(w, templBuffer)
			}
			return err
		})
		err = ButtonLink("/", "inline-block mt-6 py-4 md:text-base").Render(templ.WithChildren(ctx, var_5), templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div></div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}
