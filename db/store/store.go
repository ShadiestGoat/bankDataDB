// DONT EDIT: Auto generated

package store

import (
	"context"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/utils/erriter"
)

// Store ...
type Store interface {
	NewCategory(ctx context.Context, authorID string, name string, icon string, color string) (string, error)
	ExtDelCategory(ctx context.Context, authorID string, iD string) (int64, error)
	ExtGetCategories(ctx context.Context, authorID string) ([]*ExtGetCategoriesRow, error)
	TransMapsRmCategories(ctx context.Context, mappingID string) error
	TransMapsRmNames(ctx context.Context, mappingID string) error
	DoesCategoryExist(ctx context.Context, authorID string, iD string) (bool, error)
	DoesMappingExist(ctx context.Context, authorID string, iD string) (bool, error)
	DoesTransactionExist(ctx context.Context, authorID string, authedAt time.Time, settledAt time.Time, description string, amount float64) (bool, error)
	GetTransCount(ctx context.Context, authorID string) (int64, error)
	GetUserUpdatedAt(ctx context.Context, id string) (time.Time, error)
	ResetCategoryData(ctx context.Context, iD string, name string, color string, icon string) error
	SendBatch(ctx context.Context, b *pgx.Batch) error
	TxFunc(ctx context.Context, h func(s Store) error) error
	MappingGetAll(ctx context.Context, authorID string) ([]*data.Mapping, error)
	MappingInsert(ctx context.Context, authorID string, m *data.Mapping) (string, error)
	TransMapsMapExisting(ctx context.Context, updateName bool, newVal any, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (*erriter.Iter[string], int, error)
	TransMapsInsert(ctx context.Context, transIDs erriter.Iter[string], mappingID string, mappedName bool) error
	TransMapsInsertBatch(ctx context.Context, b *TransMapsBatch) error
	InsertCheckpoint(batch *pgx.Batch, date time.Time, amt float64)
	InsertTransactions(ctx context.Context, b *TransactionBatch) (int64, error)
	GetTransactions(ctx context.Context, authorID string, amount, offset int, orderColumn string, asc bool) ([]*data.Transactions, error)
	GetUserByName(ctx context.Context, name string) (*User, error)
	// Create a user in the DB
	// Returns the ID & an err
	// The password should be encrypted
	NewUser(ctx context.Context, username string, password []byte) (string, error)
}
