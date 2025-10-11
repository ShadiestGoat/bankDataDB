package external

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

// For /transactions
func mountTransactions(a *internal.API, r chi.Router) {
	defHTTPPage(
		r, a,
		[]string{"authed_at", "settled_at", "amount", "category"},
		func(r *http.Request, size, off int, orderBy string, asc bool) (any, int, errors.GenericHTTPError) {
			res, err := a.GetTransactions(r.Context(), getUserID(r), size, off, internal.TransactionOrderBy(orderBy), asc)
			if err != nil {
				return nil, 0, errors.InternalErr
			}

			return res, 0, nil
		},
	)
}
