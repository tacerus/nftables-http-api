package server

import (
	"log/slog"
	"net/http"
)

func (app *App) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/element/{nfFamily}/{nfTable}/{nfSet}", app.elementHandler)

	return mux
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
