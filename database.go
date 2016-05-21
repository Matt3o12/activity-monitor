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
