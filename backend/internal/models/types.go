package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONB represents a JSONB database type that can be used with PostgreSQL
type JSONB json.RawMessage

// Value implements the driver.Valuer interface for JSONB
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// Scan implements the sql.Scanner interface for JSONB
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*j = JSONB(v)
		return nil
	case string:
		*j = JSONB([]byte(v))
		return nil
	default:
		return errors.New("cannot scan unknown type into JSONB")
	}
}

// MarshalJSON implements json.Marshaler for JSONB
func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler for JSONB
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		return errors.New("models.JSONB: UnmarshalJSON on nil pointer")
	}
	*j = JSONB(data)
	return nil
}