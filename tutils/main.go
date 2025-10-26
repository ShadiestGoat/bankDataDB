package tutils

import (
	"testing"
	"context"

	"github.com/shadiestgoat/bankDataDB/db"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/db/store/mock_store"
	"github.com/shadiestgoat/bankDataDB/internal"
	"github.com/shadiestgoat/bankDataDB/log"
	"github.com/stretchr/testify/mock"
)

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

func MockStoreTx(t *testing.T, s *mock_store.MockStore) *mock_store.MockStore {
	tx := mock_store.NewMockStore(t)

	s.EXPECT().TxFunc(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, h func(s store.Store) error) error {
		return h(tx)
	})

	return tx
}