package property

import (
	"encoding/json"
	"fmt"
)

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
		return property, json.Unmarshal(data, &property)
	case "created_time":
		var property CreatedTime
		return property, json.Unmarshal(data, &property)
	case "checkbox":
		var property Checkbox
		return property, json.Unmarshal(data, &property)
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
