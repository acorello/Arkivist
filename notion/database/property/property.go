// [doc] https://developers.notion.com/reference/property-object
package property

import (
	"encoding/json"
	"fmt"
)

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

type PropertiesByName map[string]Property

func (I PropertiesByName) UnmarshalJSON(data []byte) error {
	var jsonMap rawJSONMap
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return err
	}
	for name, data := range jsonMap {
		var m rawJSONMap
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		typ, err := m.decodeStringEntry("type")
		if err != nil {
			return fmt.Errorf("failed to read type of %q: %v", name, err)
		}
		property, err := decodeProperty(data, typ)
		if err != nil {
			return fmt.Errorf("failed to decode property %q of type %q: %v", name, typ, err)
		}
		I[name] = property
	}
	return nil
}

func decodeProperty(data []byte, typ string) (Property, error) {
	switch typ {
	case "select":
		var property Select
		if err := json.Unmarshal(data, &property); err != nil {
			return nil, err
		}
		return property, nil
	case "created_time":
		var property CreatedTime
		if err := json.Unmarshal(data, &property); err != nil {
			return nil, err
		}
		return property, nil
	case "checkbox":
		var property Checkbox
		if err := json.Unmarshal(data, &property); err != nil {
			return nil, err
		}
		return property, nil
	default:
		return nil, fmt.Errorf("type %q not implemented", typ)
	}
}

type rawJSONMap map[string]json.RawMessage

func (I rawJSONMap) decodeEntry(key string, target any) error {
	rawValue, found := I[key]
	if !found {
		return fmt.Errorf("map key not found: %q", key)
	}
	if err := json.Unmarshal(rawValue, target); err != nil {
		return err
	}
	return nil
}

func (I rawJSONMap) decodeStringEntry(key string) (res string, err error) {
	err = I.decodeEntry(key, &res)
	return
}
