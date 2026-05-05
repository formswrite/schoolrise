package customfields

import "encoding/json"

type FieldType string

const (
	FieldTypeText     FieldType = "text"
	FieldTypeNumber   FieldType = "number"
	FieldTypeDate     FieldType = "date"
	FieldTypeBool     FieldType = "bool"
	FieldTypeSelect   FieldType = "select"
	FieldTypeMulti    FieldType = "multi"
)

type Definition struct {
	ID         int64           `json:"id"`
	Entity     string          `json:"entity"`
	Code       string          `json:"code"`
	Label      string          `json:"label"`
	Type       FieldType       `json:"type"`
	Required   bool            `json:"required"`
	Options    json.RawMessage `json:"options,omitempty"`
	Order      int             `json:"order"`
}

type Value struct {
	DefinitionID int64           `json:"definition_id"`
	EntityID     int64           `json:"entity_id"`
	Data         json.RawMessage `json:"data"`
}
