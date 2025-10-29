package external

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

func routeCategories(r chi.Router, a *internal.API, store store.Store) {
	defHTTP(r, `GET`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
		data, err := store.ExtGetCategories(r.Context(), getUserID(r))
		if err != nil {
			return nil, errors.InternalErr
		}

		return data, nil
	})
	defHTTPRead(r, `POST`, `/`, a, func(r *http.Request, b internal.SavableCategory) (any, errors.GenericHTTPError) {
		id, err := a.CreateCategory(r.Context(), &b, getUserID(r))
		if err != nil {
			if err, ok := err.(*internal.ValidationErr); ok {
				return nil, &errors.OnDemandHTTPError{
					Status: 400,
					Message: "Failed to validate category",
					Details: err.Details,
				}
			}
			return nil, errors.InternalErr
		}

		return &RespCreated{id}, nil
	})

	r.Route(`/{id}`, func(r chi.Router) {
		r.Use(makeMiddleware(a, func(a *internal.API, r *http.Request) (*http.Request, error) {
			exist, err := store.DoesCategoryExist(r.Context(), getUserID(r), chi.URLParam(r, "id"))
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.NotFound
			}

			return nil, nil
		}))

		defHTTPRead(r, `PUT`, `/`, a, func(r *http.Request, b internal.SavableCategory) (any, errors.GenericHTTPError) {
			err := a.UpdateCategory(r.Context(), chi.URLParam(r, "id"), &b)
			if err != nil {
				if err, ok := err.(*internal.ValidationErr); ok {
					return nil, &errors.OnDemandHTTPError{
						Status: 400,
						Message: "Failed to validate category",
						Details: err.Details,
					}
				}
				return nil, errors.InternalErr
			}

			return nil, nil
		})

		defHTTP(r, `DELETE`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
			_, err := store.ExtDelCategory(r.Context(), getUserID(r), chi.URLParam(r, "id"))
			if err != nil {
				return nil, errors.InternalErr
			}

			return nil, nil
		})
	})
}
