package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/nftables"
	"golang.org/x/sys/unix"

	nftapi "github.com/tacerus/nftables-http-api/nftables"
)

func (app *App) setGet(w http.ResponseWriter, r *http.Request, table *nftables.Table, set *nftables.Set, elements []string) {
	if table == nil {
		app.errorHandler(w, http.StatusNotFound, "Table not found")
		return
	}

	if set == nil {
		app.errorHandler(w, http.StatusNotFound, "Set not found")
		return
	}

	out := &nftapi.Set{
		Elements: elements,
		Flags:    nftapi.GetSetFlags(set),
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

func (app *App) setPut(w http.ResponseWriter, r *http.Request, nft *nftables.Conn, table *nftables.Table, set *nftables.Set, elements []string) {
	if table == nil {
		app.errorHandler(w, http.StatusNotFound, "Table not found")
		return
	}

	if set != nil {
		nft.DelSet(set)
	}

	apiSet := new(nftapi.Set)
	err := json.NewDecoder(r.Body).Decode(apiSet)
	if err != nil {
		slog.Debug("Failed to decode client payload.", "error", err)
		app.errorHandler(w, http.StatusBadRequest, "Invalid payload.")
		return
	}

	err = nftapi.AddSet(nft, table, *apiSet)
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Set creation failed.")
		return
	}

	err = nft.Flush()
	if err != nil {
		slog.Debug("Failed to flush.", "operation", "setPut", "error", err)
		app.errorHandler(w, http.StatusInternalServerError, "Flushing of set creation failed.")
		return
	}

	j, err := json.Marshal(genericOut{
		Message: "ok",
	})

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Fatal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(j)
}

func (app *App) setHandler(w http.ResponseWriter, r *http.Request) {
	nft, err := nftapi.Connect()
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to initialize nftables")
		return
	}

	familyName := r.PathValue("nfFamily")
	tableName := r.PathValue("nfTable")
	setName := r.PathValue("nfSet")

	var table *nftables.Table
	var set *nftables.Set
	var elements []string

	table, err = nftapi.GetTable(nft, familyName, tableName)
	if err != nil {
		if errors.Is(err, nftapi.ErrUnknownFamily) {
			app.errorHandler(w, http.StatusBadRequest, "Specified family is not valid.")
			return
		}

		app.errorHandler(w, http.StatusInternalServerError, "Failed to get tables")
		return
	}

	if table == nil {
		goto processSet
	}

	set, err = nftapi.GetSet(nft, table, setName)
	if err != nil && !errors.Is(err, unix.ENOENT) {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to get sets")
		return
	}

	if set == nil {
		goto processSet
	}

	elements, err = nftapi.GetSetElements(nft, set)
	if err != nil {
		app.errorHandler(w, http.StatusInternalServerError, "Failed to get elements")
		return
	}

processSet:
	switch r.Method {
	case http.MethodGet:
		app.setGet(w, r, table, set, elements)
	case http.MethodPut:
		app.setPut(w, r, nft, table, set, elements)
	}
}
