package main

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"gopkg.in/pg.v4"
)

const databaseConfigName = "database-config.json"

// DatabaseConfig contains all configuration values for the
// database.
type DatabaseConfig struct {
	User     string `json: user`
	Address  string `json: address`
	Password string `json: password`
	Database string `json: database`
	SSL      bool   `json: ssl`
}

// ToPGOptions returns a pg.Options object for the config.
func (c DatabaseConfig) ToPGOptions() *pg.Options {
	return &pg.Options{
		Addr:     c.Address,
		Password: c.Password,
		Database: c.Database,
		SSL:      c.SSL,
		User:     c.User,
	}
}

// LoadDatabaseConfig loads the config from disk or returns
// an error
func LoadDatabaseConfig() (DatabaseConfig, error) {
	config := DatabaseConfig{
		SSL:      true,
		Address:  "localhost",
		User:     "postgres",
		Database: "postgres",
	}
	content, err := ioutil.ReadFile(databaseConfigName)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(content, &config)
	return config, err
}

// Connects to the database.
func InitConnection() *pg.DB {
	config, err := LoadDatabaseConfig()
	if err != nil {
		log.Printf("Error loading database config: %v", err)
	}

	return pg.Connect(config.ToPGOptions())
}

// TransactionErrorHandler is used for dealing with
// mutiple errors during a transaction without
// checking for an error every time.
type TransactionErrorHandler struct {
	err HTTPError
}

// Err saves the error if err is not nil as a
// DatabaseError. Make sure that the error is in
// fact a database related error (i.e. it has been
// returned by the pg library).
// It returns itself so it can be chained.
func (h *TransactionErrorHandler) Err(err error) *TransactionErrorHandler {
	if h.err == nil {
		h.err = NewDatabaseError(err)
	}

	return h
}

// FirstErr returns the first error to occur during
// the transaction or nil if none has occurred.
// When an error occurred during a database operation,
// only the first one is relevant (the other ones will
// most likely only state that the database has been
// reset).
func (h *TransactionErrorHandler) FirstErr() HTTPError {
	return h.err
}
