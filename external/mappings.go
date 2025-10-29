package external

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

type RespCreatedMapping struct {
	ID string `json:"id"`
	MappingsResolved int `json:"mappingsResolved"`
}

func getRetroactive(r *http.Request) bool {
	return r.URL.Query().Get("no_retroactive") != "1"
}

func getMapping(r *http.Request) *data.Mapping {
	return r.Context().Value(CTX_MAPPING).(*data.Mapping)
}


func routeMappings(r chi.Router, a *internal.API, store store.Store) {
	defHTTP(r, `GET`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
		data, err := store.MappingGetAll(r.Context(), getUserID(r))
		if err != nil {
			return nil, errors.InternalErr
		}

		return data, nil
	})

	defHTTPRead(r, `POST`, `/`, a, func(r *http.Request, b data.Mapping) (any, errors.GenericHTTPError) {
		id, resolved, err := a.MappingCreate(r.Context(), getUserID(r), &b, getRetroactive(r))
		if err != nil {
			if err, ok := err.(*internal.ValidationErr); ok {
				return nil, &errors.OnDemandHTTPError{
					Status: 400,
					Message: "Failed to validate mapping",
					Details: err.Details,
				}
			}
			return nil, errors.InternalErr
		}

		return &RespCreatedMapping{id, resolved}, nil
	})

	r.Route(`/{id}`, func(r chi.Router) {
		r.Use(makeMiddleware(a, func(a *internal.API, r *http.Request) (*http.Request, error) {
			m, err := store.MappingGetByID(r.Context(), getUserID(r), chi.URLParam(r, "id"))
			if err != nil {
				return nil, err
			}
			if m == nil {
				return nil, errors.NotFound
			}

			return r.WithContext(context.WithValue(r.Context(), CTX_MAPPING, m)), nil
		}))

		defHTTPRead(r, `PUT`, `/`, a, func(r *http.Request, b data.Mapping) (any, errors.GenericHTTPError) {
			err := a.MappingUpdate(
				r.Context(),
				getUserID(r), getMapping(r), &b,
				getRetroactive(r),
			)

			if err != nil {
				if err, ok := err.(*internal.ValidationErr); ok {
					return nil, &errors.OnDemandHTTPError{
						Status: 400,
						Message: "Failed to validate mapping",
						Details: err.Details,
					}
				}

				return nil, errors.InternalErr
			}

			return nil, nil
		})

		defHTTP(r, `DELETE`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
			err := a.MappingDelete(r.Context(), chi.URLParam(r, "id"), getRetroactive(r))
			if err != nil {
				return nil, errors.InternalErr
			}

			return nil, nil
		})
	})
}
