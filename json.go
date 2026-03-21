package grantprovider

import (
	"encoding/json"
	"fmt"
	"io"
)

// ToJSON escribe v como JSON en w, sin escapar HTML en strings.
func ToJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// FromJSON decodifica JSON desde data en T. Rechaza campos JSON no declarados
// en el tipo destino (DisallowUnknownFields).
func FromJSON[T any](data io.Reader) (T, error) {
	var v T
	decoder := json.NewDecoder(data)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&v); err != nil {
		return v, fmt.Errorf("failed to decode JSON: %w", err)
	}
	return v, nil
}
