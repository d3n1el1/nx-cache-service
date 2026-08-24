package api

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, v any) error {
	var response any

	if v == nil {
		response = map[string]string{}
	} else {
		response = v
	}

	b, err := json.Marshal(response)
	if err != nil {
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

	err := writeJSON(w, status, response)

	if err != nil {
		return err
	}

	return err
}
