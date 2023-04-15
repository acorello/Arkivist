package property

import (
	"encoding/json"
	"fmt"
)

// List of property types.
//
// Extracted from [notion ref docs] with JS
//
//	const selector = "h2 + p + p code"
//	const nodes = document.querySelectorAll(selector)
//	Array.from(nodes).map(el => el.outerText)
//
// [notion ref docs]: https://developers.notion.com/reference/property-object
const (
	checkboxTyId       TypeId = "checkbox"
	createdByTyId      TypeId = "created_by"
	createdTimeTyId    TypeId = "created_time"
	dateTyId           TypeId = "date"
	emailTyId          TypeId = "email"
	formulaTyId        TypeId = "formula"
	lastEditedByTyId   TypeId = "last_edited_by"
	lastEditedTimeTyId TypeId = "last_edited_time"
	multiSelectTyId    TypeId = "multi_select"
	optionsTyId        TypeId = "options"
	phoneNumberTyId    TypeId = "phone_number"
	relationTyId       TypeId = "relation"
	rollupTyId         TypeId = "rollup"
	selectTyId         TypeId = "select"
	urlTyId            TypeId = "url"
)

type Id string
type Name string
type TypeId string

func (I TypeId) String() string {
	return string(I)
}

func (I TypeId) unmarshal(data []byte) (Property, error) {
	switch I {
	case selectTyId:
		var property Select
		return property, json.Unmarshal(data, &property)
	case createdTimeTyId:
		var property CreatedTime
		//lint:ignore SA9005 reason:unmarshaling works correctly with embedded tiny-types (Id, Name)
		return property, json.Unmarshal(data, &property)
	case checkboxTyId:
		var property Checkbox
		//lint:ignore SA9005 reason:unmarshaling works correctly with embedded tiny-types (Id, Name)
		return property, json.Unmarshal(data, &property)
	default:
		return nil, fmt.Errorf("type %q not implemented", I)
	}
}

type Checkbox struct {
	Id
	Name
}

func (I Checkbox) TypeId() TypeId {
	return checkboxTyId
}

type Select struct {
	Id
	Name
	Data SelectData `json:"select"`
}

func (I Select) TypeId() TypeId {
	return selectTyId
}

type SelectData struct {
	Options []SelectOption
}

type SelectOption struct {
	Id
	Name
	Color string
}

type CreatedTime struct {
	Id
	Name
}

func (I CreatedTime) TypeId() TypeId {
	return createdTimeTyId
}
