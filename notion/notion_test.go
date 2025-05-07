package notion_test

import (
	"os"
	"testing"

	"dev.acorello.it/go/arkivist/notion"
	"dev.acorello.it/go/arkivist/notion/database"
	"github.com/stretchr/testify/assert"
)

var api notion.Notion

func TestEnv(t *testing.T) {
	// guess I was just checking how to get the env? was I running any useful stuff from tests?
	const ENVK = "NOTION_AUTH_TOKEN"
	val, found := os.LookupEnv(ENVK)
	if !found {
		t.Error(ENVK, "not found")
	} else {
		println(ENVK, "=", val)
	}
}

func Test_GetDatabase(t *testing.T) {
	t.Skip("SKIP REASON: Exploratory test that calls a third-party service")
	const booksDBId = "d4f844f4-e08d-4113-9ec1-131fa6624399"
	db, err := api.Database(booksDBId)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, database.Id(booksDBId), db.Id)
}
