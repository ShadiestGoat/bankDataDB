// DONT EDIT: Auto generated

package store

import (
	"context"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/data"
)

// Store ...
type Store interface {
	CreateCategory(ctx context.Context, authorID string, name string, icon string, color string) (string, error)
	ExtDelCategory(ctx context.Context, authorID string, iD string) (int64, error)
	ExtGetCategories(ctx context.Context, authorID string) ([]*ExtGetCategoriesRow, error)
	DoesCategoryExist(ctx context.Context, authorID string, iD string) (bool, error)
	DoesTransactionExist(ctx context.Context, authorID string, authedAt time.Time, settledAt time.Time, description string, amount float64) (bool, error)
	GetTransCount(ctx context.Context, authorID string) (int64, error)
	GetUserUpdatedAt(ctx context.Context, id string) (time.Time, error)
	SendBatch(ctx context.Context, b *pgx.Batch) error
	UpdateTransCatsUsingMapping(ctx context.Context, newCategoryID string, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (int, error)
	UpdateTransNamesUsingMapping(ctx context.Context, newName string, authorID string, amtMatcher *float64, txtMatcher *regexp.Regexp) (int, error)
	GetMappingsForAuthor(ctx context.Context, authorID string) ([]*data.Mapping, error)
	InsertCheckpoint(batch *pgx.Batch, date time.Time, amt float64)
	InsertTransactions(ctx context.Context, b *TransactionBatch) (int64, error)
	GetTransactions(ctx context.Context, authorID string, amount, offset int, orderColumn string, asc bool) ([]*data.Transactions, error)
	GetUserByName(ctx context.Context, name string) (*User, error)
	// Create a user in the DB
	// Returns the ID & an err
	// The password should be encrypted
	NewUser(ctx context.Context, username string, password []byte) (string, error)
}
