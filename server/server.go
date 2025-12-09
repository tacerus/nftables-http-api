package server

import (
	"log/slog"
	"net/http"
)

func (app *App) newMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/set/{nfFamily}/{nfTable}/{nfSet}", app.setHandler)

	return mux
}

func (app *App) Start() *http.Server {
	mux := app.newMux()
	srv := &http.Server{
		Addr:    app.bind,
		Handler: app.authHandler(mux),
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
