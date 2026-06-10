package handlers

import (
	"encoding/json"
	"net/http"
)

// decodeJSON reads and unmarshals JSON from the request body.
func decodeJSON(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
