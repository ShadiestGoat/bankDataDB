package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"

	"github.com/shadiestgoat/bankDataDB/config"
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external"
	"github.com/shadiestgoat/bankDataDB/internal"
	"github.com/shadiestgoat/bankDataDB/log"
)

func main() {
	cleanup := config.LoadBasics()
	defer cleanup()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	apiDB := db.GetDB(log.NewCtxLogger().With("parent_module", "internal"))
	a := internal.NewAPI("http_server", log.NewCtxLogger(), &internal.APIConfig{
		JWT: &internal.JWTConfig{Secret: []byte(config.JWT_SECRET)},
	}, apiDB, store.NewStore(apiDB))

	r := external.Router(a, store.NewStore(db.GetDB(log.NewCtxLogger().With("parent_module", "external"))))

	a.Logger()(context.Background()).Infow("Loading server", "port", port)

	s := &http.Server{Addr: ":" + port, Handler: r}
	closingServer := make(chan bool)

	go func() {
		err := s.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			a.Logger()(context.Background()).Errorw("Closing HTTP Server", "error", err)
		} else {
			a.Logger()(context.Background()).Debugw("Server has closed intentionally(?)")
		}

		close(closingServer)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	select {
	case <-c:
		s.Close()
	case <-closingServer:
		panic("Existing server (not happy)...")
	}
}
