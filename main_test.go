package main

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	bck := Debug
	Debug = false

	retCode := m.Run()

	Debug = bck
	os.Exit(retCode)
}
