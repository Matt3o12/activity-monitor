package main

import (
	"fmt"

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
