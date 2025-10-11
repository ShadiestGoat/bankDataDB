package utils

import (
	"testing"

	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/db/store/mock_store"
	"github.com/shadiestgoat/bankDataDB/internal"
	"github.com/shadiestgoat/bankDataDB/log"
)

func Ptr[T any](v T) *T { return &v }

func MarkDBTest(t *testing.T) {
	if !db.DBDefined() {
		t.Skip("Skipping DB Test: no DB Defined!")
	}

	t.Log("Running a DB Test")
}

func NewMockAPI(t *testing.T) (*internal.API, *mock_store.MockStore) {
	store := mock_store.NewMockStore(t)

	return internal.NewAPI(
		"testing",
		log.NewTestCtxLogger(t),
		&internal.APIConfig{
			JWT: &internal.JWTConfig{
				Secret: []byte(`Shhhh! Super secret key!`),
			},
		}, nil, store,
	), store
}
