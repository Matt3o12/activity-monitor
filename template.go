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

func checkLength(templates []string) {
	if len(templates) <= 0 {
		panic("templates must have at least one argument")
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

	// The writer to write the text to
	Writer http.ResponseWriter

	// Handler if an error occured while executing the template.
	ServerErrorHandler ServerErrorHandler

	// The HTTP Status Code
	statusCode int

	// The error, if any occured
	err *HTTPError

	// The args for executing the template
	tmplArgs interface{}
}

// Configure sets the Template and Writer and returns a copy with the
// changed values.
func (w TemplateWriter) Configure(t Template, writer http.ResponseWriter) TemplateWriter {
	w.Template = t
	w.Writer = writer

	return w
}

// SetError returns a new TemplateWriter with the given error.
func (w TemplateWriter) SetError(err HTTPError) TemplateWriter {
	w.err = &err
	w.statusCode = err.Status

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
func (w TemplateWriter) Execute() bool {
	if w.err != nil {
		return w.err.WriteToPage(w.Writer)
	}

	if w.statusCode != 0 {
		w.Writer.WriteHeader(w.statusCode)
	}

	tmpl := w.Template
	w.Writer.Header().Set("Content-Type", htmlContent)
	err := tmpl.Execute(w.Writer, w.tmplArgs)
	if err != nil {
		w.ServerErrorHandler(err, w.Writer)
	}

	return err == nil
}
