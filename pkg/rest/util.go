package rest

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

func WriteAsJSON(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(200)
	w.Header().Set("content-type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
	return
}

func WriteAsXML(w http.ResponseWriter, data interface{}) {
	w.WriteHeader(200)
	w.Header().Set("content-type", "application/xml")
	if err := xml.NewEncoder(w).Encode(data); err != nil {
		w.WriteHeader(http.StatusExpectationFailed)
		return
	}
	return
}
