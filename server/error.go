package server

import (
	"encoding/json"
	"net/http"
)

func (app *App) errorHandlerProp(w http.ResponseWriter, status int, err interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(err)
	w.WriteHeader(status)
}

func (app *App) errorHandler(w http.ResponseWriter, status int, message string) {
	j, err := json.Marshal(genericOut{
		Message: message,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Fatal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(j)
}
