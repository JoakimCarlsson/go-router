package openapi

import (
	"encoding/json"
	"io"
)

// WriteJSON writes a JSON representation of the value to the writer
func WriteJSON(w io.Writer, value interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	return encoder.Encode(value)
}
