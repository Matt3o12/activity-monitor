package main

import (
	"errors"
	"io/ioutil"
	"testing"
)

// InitTestConnection creates a new connection
// to the database using the provided config
// file and sets the search_path to pg_temp
// so no permanent changes will be made.
// It will also patch environment variables,
// so don't forget to call the deter function
// that is returned.
func InitTestConnection(t *testing.T) func() {
	MarkLong(t)
	dbBck := db
	db = InitConnection()
	_, err := db.Exec("SELECT set_config('search_path', 'pg_temp', false);")

	deferFunc := func() {
		err := db.Close()
		db = dbBck
		if err != nil {
			t.Fatalf("Error closing the database connection: %v", err)
		}
	}

	empty := func() {}
	if err != nil {
		t.Fatalf("Database error: %v", err)
		deferFunc()
		return empty
	}

	raw, err := ioutil.ReadFile("scheme.sql")
	if err != nil {
		t.Fatalf("Error reading schema file: %v", err)
		deferFunc()
		return empty
	}

	_, err = db.Exec(string(raw))
	if err != nil {
		t.Fatalf("Database error while importing schema: %v", err)
		deferFunc()
		return empty
	}

	return deferFunc
}

func TestTransactionErrorHandler(t *testing.T) {
	h := TransactionErrorHandler{}
	assertNoError := func() {
		if h.FirstErr() != nil {
			msg := "TransactionErrorHandler recorded an error: %v"
			t.Errorf(msg, h.FirstErr())
		}
	}

	assertNoError()
	h.Err(nil)
	assertNoError()

	assertError := func(msg string) {
		if h.FirstErr() == nil || h.FirstErr().Error() != msg {
			t.Errorf("FirstErr() => %v, wanted: %v", h.FirstErr(), msg)
		}
	}
	h.Err(errors.New("foo"))
	assertError("DatabaseError: foo")
	h.Err(errors.New("bar"))
	assertError("DatabaseError: foo")
}
