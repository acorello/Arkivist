package notion_test

import (
	"testing"

	"dev.acorello.it/go/arkivist/notion"
	"dev.acorello.it/go/arkivist/notion/database"
	"github.com/stretchr/testify/assert"
)

var api notion.Notion = notion.LoadTOMLConfig("../_tmp/arkivist.secrets.toml")

func Test_GetDatabase(t *testing.T) {
	t.Skip("SKIP REASON: Exploratory test that calls a third-party service")
	const booksDBId = "d4f844f4e08d41139ec1131fa6624399"
	db, err := api.Database(booksDBId)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, database.Id("d4f844f4-e08d-4113-9ec1-131fa6624399"), db.Id)
}
