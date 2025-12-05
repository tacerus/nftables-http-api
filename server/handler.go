package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/tacerus/nftables-http-api/nftables"
)

func (app *App) elementHandler(w http.ResponseWriter, r *http.Request) {
	nft, err := nftables.Connect()
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to initialize nftables")
		return
	}

	familyName := r.PathValue("nfFamily")
	tableName := r.PathValue("nfTable")
	setName := r.PathValue("nfSet")

	table, err := nftables.GetTable(nft, familyName, tableName)
	if err != nil {
		if errors.Is(err, nftables.ErrUnknownFamily) {
			app.errorHandler(w, http.StatusBadRequest, "Specified family is not valid.")
			return
		}

		app.errorHandler(w, http.StatusInternalServerError, "Failed to get tables")
		return
	}
	if table == nil {
		app.errorHandler(w, http.StatusNotFound, "Table not found")
		return
	}

	set, err := nftables.GetSet(nft, table, setName)
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to get sets")
		return
	}
	if set == nil {
		app.errorHandler(w, http.StatusNotFound, "Set not found")
		return
	}

	elements, err := nftables.GetSetElements(nft, set)
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to get elements")
		return
	}

	out := &setOut{
		Elements: elements,
		Flags:    nftables.GetSetFlags(set),
		Name:     set.Name,
		Type:     set.KeyType.Name,
	}

	j, err := json.Marshal(out)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Fatal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)

	return
}
