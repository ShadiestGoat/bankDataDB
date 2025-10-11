package internal

import (
	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/log"
)

type API struct {
	log log.CtxLogger
	db  db.DBQuerier
	store store.Store
	cfg *APIConfig
}

type APIConfig struct {
	JWT *JWTConfig
}

// Util function for outsides
// Internal should use log
func (a API) Logger() log.CtxLogger {
	return a.log
}

func NewAPI(source string, logger log.CtxLogger, cfg *APIConfig, db db.DBQuerier, store store.Store) *API {
	return &API{
		log: logger.With("module", "api (internal)", "source", source),
		db:  db,
		store: store,
		cfg: cfg,
	}
}
