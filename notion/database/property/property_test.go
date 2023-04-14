package property_test

import (
	"encoding/json"
	"testing"

	"dev.acorello.it/go/arkivist/notion/database/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mapy = map[string]any

const (
	createdTimeName = "CreatedTime"
	checkboxName    = "CheckboxName"
	selectName      = "Food Group"
)

func propMapy(id, name, typ string, data mapy) mapy {
	return mapy{
		"id":   id,
		"name": name,
		"type": typ,
		typ:    data,
	}
}

var propertiesData = mapy{
	checkboxName:    propMapy("CB1", checkboxName, "checkbox", mapy{}),
	createdTimeName: propMapy("CT1", createdTimeName, "created_time", mapy{}),
	selectName: propMapy("SEL_1", selectName, "select", mapy{
		"options": []mapy{
			{
				"id":    "UUID-1",
				"name":  "Vegetable",
				"color": "purple",
			},
			{
				"id":    "UUID-2",
				"name":  "Fruit",
				"color": "red",
			},
		},
	}),
}

func TestPropertiesUnmarshalling(t *testing.T) {
	sampleJSON, err := json.Marshal(propertiesData)
	require.NoError(t, err)
	propertiesByName := property.PropertiesByName{}
	require.NoError(t, json.Unmarshal(sampleJSON, &propertiesByName))
	expected := property.PropertiesByName{
		createdTimeName: property.CreatedTime{
			Id:   "CT1",
			Name: createdTimeName,
		},
		checkboxName: property.Checkbox{
			Id:   "CB1",
			Name: checkboxName,
		},
		selectName: property.Select{
			Id:   "SEL_1",
			Name: selectName,
			Data: property.SelectData{
				Options: []property.SelectOption{
					{Id: "UUID-1", Name: "Vegetable", Color: "purple"},
					{Id: "UUID-2", Name: "Fruit", Color: "red"},
				},
			},
		},
	}
	assert.Equal(t, expected, propertiesByName)
}
