package property_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"dev.acorello.it/go/arkivist/notion/database/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

// How to deserialize a value to one of a set of constants?

func TestCreatedTimeDecoding(t *testing.T) {
	var createdTime property.CreatedTime
	const createdTimeJSON = `{
		"id": "CT1",
		"name": "CreatedTime",
		"type": "created_time",
		"created_time": {}
	}`
	// do I want tiny-types for each field?
	// they are definitely logically distinct sets
	//
	// is adding tiny-types later a lot of work?
	var expected = property.CreatedTime{
		Id:   "CT1",
		Name: "CreatedTime",
	}
	decoder := json.NewDecoder(strings.NewReader(createdTimeJSON))
	err := decoder.Decode(&createdTime)
	assert.NoError(t, err)
	assert.Equal(t, createdTime, expected)
	// testing this from a foreign package adds text in case of multiple asserts

	// A1: assert.Equal(t, wholeActual, wholeExpected)
	// A2: use tiny-types for everything
	//     wholeActual embeds a struct
	//        constructing it looks cumbersome (even more)
}

type Property interface {
	TypeId() string
}

type Checkbox struct {
	Id      string
	Name    string
	Checked bool
}

func (I Checkbox) TypeId() string {
	return "checkbox"
}

type Select struct {
	Id   string
	Name string
	Data SelectData
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

type PropertiesByName map[string]Property

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

func (me rawJSONMap) decodeAsSelect() (Select, error) {
	var property Select
	const fields = 3
	ks := [fields]string{"id", "name", "select"}
	ps := [fields]any{&property.Id, &property.Name, &property.Data}
	for i := 0; i < fields; i++ {
		if err := me.decodeEntry(ks[i], ps[i]); err != nil {
			return property, err
		}
	}
	return property, nil
}

func (I PropertiesByName) UnmarshalJSON(data []byte) error {
	var m rawJSONMap
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	for name, rawValue := range m {
		var property Property
		property, err := decodeProperty(rawValue)
		if err != nil {
			return fmt.Errorf("failed to decode %q: %v", name, err)
		}
		I[name] = property
	}
	return nil
}

func decodeProperty(data []byte) (Property, error) {
	var m rawJSONMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	typ, err := m.decodeStringEntry("type")
	if err != nil {
		return nil, fmt.Errorf("failed to %q: %v", typ, err)
	}
	switch typ {
	case "select":
		property, err := m.decodeAsSelect()
		if err != nil {
			return nil, err
		}
		return property, nil
	default:
		panic("N/I")
	}
}

func TestPropertiesUnmarshalling(t *testing.T) {
	const sample = `{
		"Food Group": {
			"id": "SEL_1",
			"name": "Food Group",
			"type": "select",
			"select": {
				"options": [
					{
						"id": "UUID-1",
						"name": "Vegetable",
						"color": "purple"
					},
					{
						"id": "UUID-2",
						"name": "Fruit",
						"color": "red"
					}
				]
			}
		}
	}`

	ps := PropertiesByName{}
	require.NoError(t, json.Unmarshal([]byte(sample), &ps))
	assert.ElementsMatch(t, maps.Keys(ps), []string{"Food Group"})
	assert.Equal(t, Select{
		Id:   "SEL_1",
		Name: "Food Group",
		Data: SelectData{
			Options: []SelectOption{
				{Id: "UUID-1", Name: "Vegetable", Color: "purple"},
				{Id: "UUID-2", Name: "Fruit", Color: "red"},
			},
		},
	}, ps["Food Group"])
}
