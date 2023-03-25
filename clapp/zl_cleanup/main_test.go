package main

import (
	"fmt"
	"testing"
)

func TestRun(t *testing.T) {
	config := populateConfig()
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
	s = "9781003311973 • Mastering Visual Studio Code A Beginner’s Guide • by Sufyan Bin Uzayr • CRC Press (z-lib.org)"
	s = "Mastering Visual Studio Code A Beginner’s Guide • by Sufyan Bin Uzayr • CRC Press (z-lib.org).pdf"
}

func TestAppendNil(t *testing.T) {
	errs := []error{fmt.Errorf("one")}
	for _, err := range []error{nil, nil} {
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 1 {
		t.Errorf("Append added something: %#v", errs)
	}
}
