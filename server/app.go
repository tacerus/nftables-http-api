package server

import (
	"context"

	"github.com/tacerus/nftables-http-api/core"
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
