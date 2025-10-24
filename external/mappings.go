package external

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

func routeMappings(r chi.Router, a *internal.API, store store.Store) {
	defHTTP(r, `GET`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
		data, err := store.GetMappingsForAuthor(r.Context(), getUserID(r))
		if err != nil {
			return nil, errors.InternalErr
		}

		return data, nil
	})

	defHTTPRead(r, `POST`, `/`, a, func(r *http.Request, b data.Mapping) (any, errors.GenericHTTPError) {
		id, err := a.CreateMapping(r.Context(), getUserID(r), &b)
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

		if r.URL.Query().Get("no_retroactive") != "1" {
			_, err := a.UpdateAnyMatchedMappings(r.Context(), &b, getUserID(r))
			if err != nil {
				return nil, errors.InternalErr
			}
		}

		return &RespCreated{id}, nil
	})

	r.Route(`/{id}`, func(r chi.Router) {
		r.Use(makeMiddleware(a, func(a *internal.API, r *http.Request) (*http.Request, error) {
			exist, err := store.DoesMappingExist(r.Context(), getUserID(r), chi.URLParam(r, "id"))
			if err != nil {
				return nil, err
			}
			if !exist {
				return nil, errors.NotFound
			}

			return nil, nil
		}))
	})
}
