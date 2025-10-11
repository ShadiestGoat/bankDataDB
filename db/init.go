package db

import (
	"context"
	"sync"
	"testing"
	"time"

	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shadiestgoat/bankDataDB/log"
)

var (
	globalPool     *pgxpool.Pool
	globalLoadLock = &sync.Mutex{}
)

//go:embed schema/*
var schema embed.FS

func LoadPool(uri string) {
	globalLoadLock.Lock()
	defer globalLoadLock.Unlock()

	if globalPool != nil {
		return
	}

	schemaFiles, err := schema.ReadDir("schema")
	if err != nil {
		panic("Can't read schema dir: " + err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(context.Background(), uri)
	if err != nil {
		panic("Can't load pool! " + err.Error())
	}

	err = pool.Ping(ctx)
	if err != nil {
		panic("Can't ping DB: " + err.Error())
	}

	for _, f := range schemaFiles {
		sch, err := schema.ReadFile("schema/" + f.Name())
		if err != nil {
			panic("Can't read the schema file name " + f.Name() + ": " + err.Error())
		}

		_, err = pool.Exec(context.Background(), string(sch))
		if err != nil {
			panic("Can't run migration " + f.Name() + ": " + err.Error())
		}
	}

	globalPool = pool
}

func GetDB(logger log.CtxLogger) DBQuerier {
	if !DBDefined() {
		panic("No DB Loaded!")
	}

	return &genericDBWithLog[*pgxpool.Pool]{
		conn: globalPool,
		log:  logger.With("module", "database"),
	}
}

func GetTestDB(t *testing.T) DBQuerier {
	return GetDB(log.NewTestCtxLogger(t))
}

func DBDefined() bool {
	globalLoadLock.Lock()
	defer globalLoadLock.Unlock()
	return globalPool != nil
}

func Close() {
	globalPool.Close()
}
