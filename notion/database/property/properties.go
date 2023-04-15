package property

import (
	"encoding/json"
	"fmt"
)

type Property interface {
	TypeId() TypeId
}

type PropertiesByName map[string]Property

func (I PropertiesByName) UnmarshalJSON(data []byte) error {
	var rawPropertiesByName mapOfRawJSON
	if err := json.Unmarshal(data, &rawPropertiesByName); err != nil {
		return err
	}
	for name, bytes := range rawPropertiesByName {
		property, err := decodeProperty(bytes)
		if err != nil {
			return fmt.Errorf("failed to decode property %q: %v", name, err)
		}
		I[name] = property
	}
	return nil
}

func decodeProperty(data []byte) (Property, error) {
	var rawParts mapOfRawJSON
	if err := json.Unmarshal(data, &rawParts); err != nil {
		return nil, err
	}
	typeId, err := rawParts.decodeString("type")
	if err != nil {
		return nil, fmt.Errorf("failed to read type: %v", err)
	}
	property, err := TypeId(typeId).decodeProperty(data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode property of type %q: %v", typeId, err)
	}
	return property, nil
}

func (I TypeId) decodeProperty(data []byte) (Property, error) {
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

type mapOfRawJSON map[string]json.RawMessage

func (I mapOfRawJSON) decodeEntry(key string, target any) error {
	rawValue, found := I[key]
	if !found {
		return fmt.Errorf("map key not found: %q", key)
	}
	if err := json.Unmarshal(rawValue, target); err != nil {
		return err
	}
	return nil
}

func (I mapOfRawJSON) decodeString(key string) (res string, err error) {
	err = I.decodeEntry(key, &res)
	return
}
