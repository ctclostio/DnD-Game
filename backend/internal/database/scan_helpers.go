package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

// RowScanner is an interface that both sql.Row and sql.Rows implement
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// JSONFieldUnmarshaler is a function that unmarshals JSON data into a specific field
type JSONFieldUnmarshaler func(data []byte) error

// ScanWithJSON is a generic helper for scanning rows that contain JSON fields
func ScanWithJSON(scanner RowScanner, dest []interface{}, jsonFields map[int]JSONFieldUnmarshaler) error {
	// Create temporary byte slices for JSON fields
	tempDest := make([]interface{}, len(dest))
	jsonData := make(map[int]*[]byte)

	for i, d := range dest {
		if _, hasJSON := jsonFields[i]; hasJSON {
			var data []byte
			jsonData[i] = &data
			tempDest[i] = &data
		} else {
			tempDest[i] = d
		}
	}

	// Scan the row
	if err := scanner.Scan(tempDest...); err != nil {
		return err
	}

	// Unmarshal JSON fields
	for i, unmarshaler := range jsonFields {
		if data := jsonData[i]; data != nil && *data != nil {
			if err := unmarshaler(*data); err != nil {
				return err
			}
		}
	}

	return nil
}

// ScanRowsGeneric is a generic helper for scanning multiple rows with a custom scanner function
func ScanRowsGeneric[T any](rows *sql.Rows, scanner func(RowScanner) (*T, error)) ([]*T, error) {
	defer func() { _ = rows.Close() }()

	var results []*T
	for rows.Next() {
		item, err := scanner(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Common JSON unmarshalers
func UnmarshalJSON(target interface{}) JSONFieldUnmarshaler {
	return func(data []byte) error {
		return json.Unmarshal(data, target)
	}
}

func UnmarshalJSONWithError(target interface{}, fieldName string) JSONFieldUnmarshaler {
	return func(data []byte) error {
		if err := json.Unmarshal(data, target); err != nil {
			return fmt.Errorf("failed to unmarshal %s: %w", fieldName, err)
		}
		return nil
	}
}

// MarshalJSONField marshals a value to JSON with error handling
func MarshalJSONField(value interface{}, fieldName string) ([]byte, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s: %w", fieldName, err)
	}
	return data, nil
}

// ExecWithJSON executes a query with JSON field marshaling
func ExecWithJSON(db interface {
	ExecRebind(string, ...interface{}) (sql.Result, error)
},
	query string, args []interface{}, jsonFields map[int]struct {
		Value interface{}
		Name  string
	}) error {

	// Marshal JSON fields
	for i, field := range jsonFields {
		data, err := MarshalJSONField(field.Value, field.Name)
		if err != nil {
			return err
		}
		args[i] = data
	}

	_, err := db.ExecRebind(query, args...)
	return err
}
