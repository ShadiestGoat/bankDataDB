package internal

import (
	"context"

	"github.com/shadiestgoat/bankDataDB/data"
)

type TransactionOrderBy string

const (
	TRANS_ORDER_AUTH   = "authedAt"
	TRANS_ORDER_SETTLE = "settledAt"
	TRANS_ORDER_AMT    = "amount"
)

func columnInputToTrans(inp TransactionOrderBy) string {
	switch inp {
	case TRANS_ORDER_AMT:
		return "amount"
	case TRANS_ORDER_AUTH:
		return "authed_at"
	case TRANS_ORDER_SETTLE:
		return "settled_at"
	}

	return "authed_at"
}

func (a *API) GetTransactions(ctx context.Context, authorID string, amount, offset int, orderBy TransactionOrderBy, asc bool) ([]*data.Transactions, error) {
	return a.store.GetTransactions(
		ctx, authorID,
		amount, offset, columnInputToTrans(orderBy), asc,
	)
}
