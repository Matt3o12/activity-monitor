package main

import (
	"errors"
	html "html/template"
	"io"
	"testing"
)

type nullTemplate struct{}

func (t *nullTemplate) Execute(w io.Writer, args interface{}) error {
	return nil
}

func stringSliceEqual(a, b []string) bool {
	if a == nil || b == nil || len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestFixTemplateNames(t *testing.T) {
	testcase := []struct {
		in, out []string
	}{
		{[]string{"test.html"}, []string{"templates/test.html"}},
		{[]string{"foo.html", "bar.html"}, []string{"templates/foo.html", "templates/bar.html"}},
	}

	for _, row := range testcase {
		fixTemplateNames(row.in)
		if !stringSliceEqual(row.in, row.out) {
			t.Errorf("fixTemplateNames() => %q; expected: %q", row.in, row.out)
		}
	}
}

func TestMustTemplate(t *testing.T) {
	msg := "Error creating template: 'test error'"

	defer func() {
		if r := recover(); r != msg {
			t.Errorf("Expected to panic: '%v'. Got: '%v'", msg, r)
		}
	}()

	err := errors.New("test error")
	MustTemplate(nil, err)
}

func TestMustTemplateSuccess(t *testing.T) {
	template := &nullTemplate{}
	got := MustTemplate(template, nil)
	if got.(*nullTemplate) != template {
		t.Errorf("Returned unexpected template: %q (wanted: %q)", got, template)
	}
}

func isTmlpLoaded(htmlTmp *html.Template, name string) bool {
	for _, tmp := range htmlTmp.Templates() {
		if tmp.Name() == name {
			return true
		}
	}

	return false
}

func assertBaseTemplate(t *testing.T, tmp Template, err error) *html.Template {
	// Since Template is really just a wrapper, there is not much testing we can do.
	if err != nil {
		t.Errorf("Could not load template. Error: %v", err)
	}

	htmlTmp, ok := tmp.(*html.Template)
	if !ok {
		t.Fatalf("Template is supposed to be html/*template.Template, got: %T", tmp)
	}

	if htmlTmp.Name() != "index.html" {
		t.Fatalf("Templated has an unexpected name: %v", htmlTmp.Name())
	}

	return htmlTmp
}

func TestNewTemplate(t *testing.T) {
	tmp, err := NewTemplate("index.html")
	htmlTmp := assertBaseTemplate(t, tmp, err)

	// We also want to know whether the layout template was loaded as well.
	if !isTmlpLoaded(htmlTmp, "layout.html") {
		t.Fatalf("'layout.html' was also supposed to be loaded, not included")
	}
}

func TestNewBareboneTemplate(t *testing.T) {
	tmpl, err := NewBareboneTemplate("index.html")
	htmlTmp := assertBaseTemplate(t, tmpl, err)

	// We need to make sure the layout template was NOT loaded.
	if isTmlpLoaded(htmlTmp, "layout.html") {
		t.Fatalf("'layout.html' has been loaded, although it was not supposed to.")
	}
}
