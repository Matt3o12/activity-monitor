package main

import (
	"fmt"
	"net/http"

	html "html/template"
	"io"
)

const baseLayoutName = "layout.html"

// fixTemplateNames appends 'templates/' to every name
// so that the template can be loaded from the templates
// directory.
func fixTemplateNames(templates []string) {
	for i := range templates {
		templates[i] = "templates/" + templates[i]
	}
}

// A simple template that can be executed.
type Template interface {
	Execute(io.Writer, interface{}) error
}

// eagerTemplate is used when Debug is true.
// It makes sure that the template is loaded
// every time the page is refreshed (instead
// of loading all template upfront).
type eagerTemplate struct {
	path []string
}

func (e *eagerTemplate) Execute(w io.Writer, args interface{}) error {
	t, err := html.ParseFiles(e.path...)
	if err != nil {
		return err
	}

	return t.Execute(w, args)
}

// NewTemplate creates a new template with all necessary included
func NewTemplate(path ...string) (Template, error) {
	path = append(path, baseLayoutName)
	return NewBareboneTemplate(path...)
}

// NewTemplate creates a new template with the given path(es) but
// without the layout template loaded.
func NewBareboneTemplate(path ...string) (Template, error) {
	fixTemplateNames(path)
	if Debug {
		return &eagerTemplate{path}, nil
	} else {
		return html.ParseFiles(path...)
	}
}

// Ensures that a template could be loaded. If not, a panic will
// be created.
func MustTemplate(t Template, err error) Template {
	if err != nil {
		panic(fmt.Sprintf("Error creating template: '%v'", err.Error()))
	}

	return t
}

// ServerErrorHandler is called if an error occured while
// executing the template.
type ServerErrorHandler func(error, http.ResponseWriter)

// TemplateWriter helps with error handling.
type TemplateWriter struct {
	// The template to parse
	Template Template

	// Handler if an error occured while executing the template.
	ServerErrorHandler ServerErrorHandler

	// The HTTP Status Code
	StatusCode int

	// The error, if any occured
	Err HTTPError

	// The args for executing the template
	TmplArgs interface{}
}

// SetTemplate returns a new TemplateWriter with the template.
func (w TemplateWriter) SetTemplate(tmpl Template) TemplateWriter {
	w.Template = tmpl
	return w
}

// SetError returns a new TemplateWriter with the given error.
func (w TemplateWriter) SetError(err HTTPError) TemplateWriter {
	if err != nil {
		w.Err = err
		w.StatusCode = err.HTTPStatus()
	}

	return w
}

// SetStatusCode returns a new TemplateWriter with the given status code.
func (w TemplateWriter) SetStatusCode(code int) TemplateWriter {
	w.StatusCode = code

	return w
}

// SetTmplArgs returns a new TemplateWriter with the given template args.
func (w TemplateWriter) SetTmplArgs(args interface{}) TemplateWriter {
	w.TmplArgs = args

	return w
}

// Execute writes the template (and status code) or error (if set)
// to the reponse writer.
func (w TemplateWriter) Execute(httpWriter http.ResponseWriter) bool {
	if w.Err != nil {
		return w.Err.WriteToPage(httpWriter)
	}

	httpWriter.Header().Set("Content-Type", htmlContent)
	if w.StatusCode != 0 {
		httpWriter.WriteHeader(w.StatusCode)
	}

	tmpl := w.Template
	_ = "breakpoint"
	err := tmpl.Execute(httpWriter, w.TmplArgs)
	if err != nil {
		w.ServerErrorHandler(err, httpWriter)
	}

	return err == nil
}
