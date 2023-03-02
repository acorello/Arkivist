package main

import "testing"

func TestRun(t *testing.T) {
	config := initConfig()
	cleanup(config)
}

func TestHasUnorthodoxRune(t *testing.T) {
	s := "9781101152140 • Drive • by Daniel H. Pink • Riverhead Books (z-lib.org)"
	if !hasUnorthodoxRune(s) {
		t.Error("Failed to match")
	}
	s = "9781101152140 • Drive • by Daniel H. Pink • Riverhead Books (z-lib.org)"
	if hasUnorthodoxRune(s) {
		t.Error("Matched too much")
	}
}
