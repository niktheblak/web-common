package response

import (
	"encoding/json"
	"net/http"
)

// Encode any Go type as JSON to the given http.ResponseWriter, while also setting the appropriate headers.
func Encode(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// Decode a JSON HTTP body into a map.
func Decode(r *http.Request) (map[string]any, error) {
	if r == nil {
		return map[string]any{}, nil
	}
	if r.Body == nil {
		return map[string]any{}, nil
	}
	defer r.Body.Close()
	dec := json.NewDecoder(r.Body)
	m := make(map[string]any)
	err := dec.Decode(&m)
	return m, err
}
