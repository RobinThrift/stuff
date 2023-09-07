package views

type Model[D any] struct {
	Global Global
	Data   D
}

type Global struct {
	Title        string
	CSRFToken    string
	FlashMessage string
	Search       string
}

type DocumentViewModel struct {
	Global Global
	Body   string
	Data   any
}

type LoginPageViewModel struct {
	ValidationErrs map[string]string
}

type ChangePasswordPageViewModel struct {
	ValidationErrs map[string]string
}
