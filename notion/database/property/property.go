// [doc] https://developers.notion.com/reference/property-object
package property

type Name string
type Id string

// List of property types.
//
// Extracted from [notion ref docs] with JS
//
//	const selector = "h2 + p + p code"
//	const nodes = document.querySelectorAll(selector)
//	Array.from(nodes).map(el => el.outerText)
//
// [notion ref docs]: https://developers.notion.com/reference/property-object
var _ = [...]string{
	"created_by",
	"created_time",
	"date",
	"email",
	"formula",
	"last_edited_by",
	"last_edited_time",
	"multi_select",
	"options",
	"phone_number",
	"relation",
	"rollup",
	"select",
	"url",
}

type Property interface {
	TypeId() string
}

type Checkbox struct {
	Id   string
	Name string
}

func (I Checkbox) TypeId() string {
	return "checkbox"
}

type Select struct {
	Id   string
	Name string
	Data SelectData `json:"select"`
}

func (I Select) TypeId() string {
	return "select"
}

type SelectData struct {
	Options []SelectOption
}

type SelectOption struct {
	Id    string
	Name  string
	Color string
}

type CreatedTime struct {
	Id   Id
	Name Name
}

func (I CreatedTime) TypeId() string {
	return "created_time"
}
