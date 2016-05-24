package main

import (
	"errors"
	html "html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestTemplateConfigure(t *testing.T) {
	handlerCalled := false
	nullHandler := func(err error, w http.ResponseWriter) {
		handlerCalled = true
	}
	tw := TemplateWriter{ServerErrorHandler: nullHandler}

	tmpl := nullTemplate{}
	new := tw.SetTemplate(&tmpl)

	if new.Template == nil {
		t.Errorf("Template not updated by Configure")
	}

	// The only way to verify ServerErrorHandler's integrity is to call
	// it and check if bahavior has been changed.
	new.ServerErrorHandler(nil, nil)
	if !handlerCalled {
		t.Errorf("ServerErrorHandler has unexpecedly been changed.")
	}

	if tw.Template != nil {
		t.Errorf("Original TemplateWriter has been changed by Configure")
	}
}

type mockedTemplate struct {
	returnError bool
	called      bool
	args        interface{}
	writer      io.Writer
}

func (t *mockedTemplate) Execute(w io.Writer, args interface{}) error {
	t.args = args
	t.writer = w
	t.called = true

	if t.returnError {
		return errors.New("Mocked template error")
	}

	return nil
}

func assertTemplate(t *testing.T, tmpl *mockedTemplate, w io.Writer, args interface{}) {
	if !tmpl.called {
		t.Errorf("Template has not been called.")
	}
	if tmpl.writer != w {
		t.Errorf("Template were not given the correct writer")
	}

	if tmpl.args != args {
		t.Errorf("Template were not given the correct args: %q", tmpl.args)
	}
}

func TestTemplateWriterExecute(t *testing.T) {
	recorder := httptest.NewRecorder()
	tmpl := &mockedTemplate{returnError: false}
	tw := TemplateWriter{
		ServerErrorHandler: nil, Template: tmpl,
	}
	ok := tw.SetStatusCode(301).SetTmplArgs("12542").Execute(recorder)

	if !ok {
		t.Error("Execution expected to be ok")
	}

	if recorder.Code != 301 {
		t.Errorf("StatusCode has not been written to the recorder.")
	}

	htmlType := "text/html; charset=utf-8"
	if c := recorder.HeaderMap.Get("Content-Type"); c != htmlType {
		t.Errorf("Template set unexpected type: %v", c)
	}
	assertTemplate(t, tmpl, recorder, "12542")
}

func TestTemplateWriterExecuteError(t *testing.T) {
	tmpl := &mockedTemplate{returnError: false}
	recorder := httptest.NewRecorder()
	tw := TemplateWriter{
		ServerErrorHandler: nil, Template: tmpl,
	}
	msg := "Page 'test' could not be found"
	httpErr := StatusError{Status: 404, Message: msg}
	tw.SetTemplate(tmpl).SetError(httpErr).Execute(recorder)

	if tmpl.called {
		t.Errorf("Mocked template was not supposed to be called.")
	}

	if recorder.Code != 404 {
		t.Errorf("Unexpected error code was set: %v", recorder.Code)
	}

	expected := "<title>Error 404 – Page " +
		"&#39;test&#39; could not be found</title>"

	if b := recorder.Body.String(); !strings.Contains(b, expected) {
		t.Errorf("%q not found in body: %q.", expected, b)
	}
}

func assertNoError(t *testing.T, tw TemplateWriter) {
	if tw.Err != nil {
		t.Errorf("Did not expect tw to have an error, got: %v", tw.Err)
	}
}

func assertError(t *testing.T, err error, tw TemplateWriter) {
	if tw.Err != err {
		t.Errorf("tw.Err => %v; wanted: %v", tw.Err, err)
	}
}

func TestTemplateWriterSetError(t *testing.T) {
	tw := TemplateWriter{}
	assertNoError(t, tw)

	tw = tw.SetError(nil)
	assertNoError(t, tw)

	testErr := NewDatabaseError(errors.New("test error"))
	tw = tw.SetError(testErr)
	assertError(t, testErr, tw)
}
