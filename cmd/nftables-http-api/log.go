package main

import (
	"context"
	"log/slog"
	"os"
)

type logHandler struct {
	*slog.JSONHandler
}

func (lh *logHandler) Handle(ctx context.Context, r slog.Record) error {
	//if val, ok := ctx.Value("xxx").(string); ok {
	//	r.AddAttrs(slog.String("xxx", val))
	//}

	return lh.JSONHandler.Handle(ctx, r)
}

func newLogLevel(inLevel string) slog.Level {
	var outLevel slog.Level
	if err := outLevel.UnmarshalText([]byte(inLevel)); err != nil {
		panic(err)
	}

	return outLevel
}

func newSlog(level slog.Level) *slog.Logger {
	return slog.New(&logHandler{
		JSONHandler: slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: level}),
	})
}
