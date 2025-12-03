package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tacerus/nftables-http-api/core"
	"github.com/tacerus/nftables-http-api/nftables"
)

type App struct {
	bind string
	Ctx  context.Context
}

func NewApp(c core.Config) *App {
	app := new(App)

	app.bind = c.Bind
	app.Ctx = context.Background()

	return app
}

func (app *App) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/element/{nfFamily}/{nfTable}/{nfSet}", app.elementHandler)

	return mux
}

type appError struct {
	Message string
}

type setOut struct {
	Elements []string
}

//func (app *App) errorHandler (w http.ResponseWriter, err interface{}) {
//    w.Header().Set("Content-Type", "application/json")
//    json.NewEncoder(w).Encode(err)
//}

func (app *App) errorHandler(w http.ResponseWriter, status int, message string) {
	j, err := json.Marshal(appError{
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

	out := new(setOut)
	out.Elements = elements

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

func (app *App) Start() *http.Server {
	mux := app.newMux()
	srv := &http.Server{
		Addr:    app.bind,
		Handler: mux,
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()

	slog.Info("Listening ...", "bind", app.bind)

	return srv
}
