package console

import (
	"html/template"
	"io"

	"github.com/labstack/echo"
)

// Template is a wrapper around all of the view templates being used in the console.
type Template struct {
	templates *template.Template
}

// Render is called to fulfill the initial page load for the dashboard.
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}
