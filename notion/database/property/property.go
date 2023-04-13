// [doc] https://developers.notion.com/reference/property-object
package property

import "encoding/json"

type Name string
type Id string

// extracted with JS
//
//	const selector = "h2 + p + p code"
//	const nodes = document.querySelectorAll(selector)
//	Array.from(nodes).map(el => el.outerText)
type TypeId string

// valid property types
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

type CreatedTime struct {
	Id   Id
	Name Name
}

func (I CreatedTime) Type() TypeId { return "created_time" }

// A: unmarshal single object to corresponding property
//
//	If I know the value in advance I can just call
//	`json.Unmarshal([]byte,any)error`.
// But what if I want to Unmarshal a larger object having one field of the interface type?

// The `checkbox` type object is empty; there is
// no additional property configuration. ([docs])
//
// [docs]: https://developers.notion.com/reference/property-object#checkbox
type Checkbox struct {
	Id
	Name
}

// forall DBs, Map has 1 `Property{type:"title"}`
type Map map[Name]Property

func (c Map) UnmarshalJSON(data []byte) error {
	var rawData map[Name]json.RawMessage
	if err := json.Unmarshal(data, &rawData); err != nil {
		return err
	}
	for n, _ := range rawData {
		// I have an JSON map with polymorphic values
		// to unmarshal a polymorphic value I have to implement custom unmarshaller
		c[n] = nil
	}
	return nil
}

// The value of type corresponds to another key in the property object. Each property object has a nested property named the same as its type value.
type Property interface {
	TypeName() TypeId
}
