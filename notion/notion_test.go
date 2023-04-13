package notion_test

import (
	"fmt"
	"testing"

	"dev.acorello.it/go/arkivist/notion"
)

var api notion.Notion = notion.LoadTOMLConfig("../_tmp/arkivist.secrets.toml")

func Test_GetDatabase(t *testing.T) {
	const booksDBId = "d4f844f4e08d41139ec1131fa6624399"
	db, err := api.Database(booksDBId)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("%v", db)
}
