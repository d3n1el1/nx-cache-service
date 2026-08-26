package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	if v == nil {
		v = struct{}{}
	}

	b, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(b)

	return err
}

func writeErrorResponse(w http.ResponseWriter, status int, reason string, additionalData any) error {
	response := map[string]any{"status": status}

	if len(reason) > 0 {
		response["reason"] = reason
	}

	if additionalData != nil {
		response["additionalData"] = additionalData
	}

	return writeJSON(w, status, response)
}
