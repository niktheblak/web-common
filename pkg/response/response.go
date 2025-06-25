package response

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Encode any Go type as JSON to the given http.ResponseWriter, while also setting the appropriate headers.
func Encode(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

// Decode a JSON HTTP body into a map.
func Decode(r *http.Request) (body map[string]any, err error) {
	body = make(map[string]any)
	if r == nil {
		return
	}
	if r.Body == nil {
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	dec := json.NewDecoder(r.Body)
	err = dec.Decode(&body)
	return
}
