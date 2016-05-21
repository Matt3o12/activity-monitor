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

// NewTemplate creates a new template with all necessary included
func NewTemplate(path ...string) (Template, error) {
	path = append(path, baseLayoutName)
	return NewBareboneTemplate(path...)
}

// NewTemplate creates a new template with the given path(es) but
// without the layout template loaded.
func NewBareboneTemplate(path ...string) (Template, error) {
	fixTemplateNames(path)
	return html.ParseFiles(path...)
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
	statusCode int

	// The error, if any occured
	err HTTPError

	// The args for executing the template
	tmplArgs interface{}
}

// SetTemplate returns a new TemplateWriter with the template.
func (w TemplateWriter) SetTemplate(tmpl Template) TemplateWriter {
	w.Template = tmpl
	return w
}

// SetError returns a new TemplateWriter with the given error.
func (w TemplateWriter) SetError(err HTTPError) TemplateWriter {
	if err != nil {
		w.err = err
		w.statusCode = err.HTTPStatus()
	}

	return w
}

// SetStatusCode returns a new TemplateWriter with the given status code.
func (w TemplateWriter) SetStatusCode(code int) TemplateWriter {
	w.statusCode = code

	return w
}

// SetTmplArgs returns a new TemplateWriter with the given template args.
func (w TemplateWriter) SetTmplArgs(args interface{}) TemplateWriter {
	w.tmplArgs = args

	return w
}

// Execute writes the template (and status code) or error (if set)
// to the reponse writer.
func (w TemplateWriter) Execute(httpWriter http.ResponseWriter) bool {
	if w.err != nil {
		return w.err.WriteToPage(httpWriter)
	}

	httpWriter.Header().Set("Content-Type", htmlContent)
	if w.statusCode != 0 {
		httpWriter.WriteHeader(w.statusCode)
	}

	tmpl := w.Template
	err := tmpl.Execute(httpWriter, w.tmplArgs)
	if err != nil {
		w.ServerErrorHandler(err, httpWriter)
	}

	return err == nil
}
