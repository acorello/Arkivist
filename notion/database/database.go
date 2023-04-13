package database

import (
	"encoding/json"
	"io"
	"time"
)

type Id string

var Zero Database = Database{}

type Database struct {
	Id             Id        `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`
	// Title          any            `json:"title"`       // TODO: specify
	// Description    any            `json:"description"` // TODO: specify
	// Properties     properties.Map `json:"properties"`  // TODO: specify
}

func ParseJSON(r io.Reader) (Database, error) {
	var db Database
	if err := json.NewDecoder(r).Decode(&db); err != nil {
		return Zero, err
	}
	return db, nil
}
