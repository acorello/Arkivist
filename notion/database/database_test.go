package database

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseJson(t *testing.T) {
	const json = "booksdb.sample.json"
	sample, err := os.Open(json)
	if err != nil {
		t.Fatal(err)
	}
	db, err := UnmarshalJSONReader(sample)
	if err != nil {
		t.Fatal(err)
	}
	expected := Database{
		Id:             "d4f844f4-e08d-4113-9ec1-131fa6624399",
		CreatedTime:    ParseTimeOrPanic("2022-07-06T12:18:00.000Z"),
		LastEditedTime: ParseTimeOrPanic("2023-04-12T08:02:00.000Z"),
	}
	assert.Equal(t, expected, db)
}

func ParseTimeOrPanic(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		log.Fatal(err)
	}
	return t
}
