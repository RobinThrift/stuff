package views

import (
	"bytes"
	"io"
)

func Render[D any](w io.Writer, name string, data Model[D]) error {
	b := bytes.NewBuffer(nil)

	if wb, ok := w.(*bytes.Buffer); ok {
		b = wb
	}

	err := execTemplate(b, name, data)
	if err != nil {
		return err
	}

	_, err = b.WriteTo(w)
	return err
}

func RenderLoginPage(w io.Writer, data Model[LoginPageViewModel]) error {
	return Render(w, "login_page", data)
}

func RenderChangePasswordPage(w io.Writer, data Model[ChangePasswordPageViewModel]) error {
	return Render(w, "change_password_page", data)
}
