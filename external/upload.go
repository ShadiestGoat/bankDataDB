package external

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

func mountUpload(api *internal.API, r chi.Router) {
	defHTTP(r, `POST`, `/`, api, func(r *http.Request) (any, errors.GenericHTTPError) {
		defer r.Body.Close()

		resp, err := api.ParseTSV(r.Context(), r.Body, getUserID(r))
		if err != nil {
			if err, ok := err.(errors.GenericHTTPError); ok {
				return nil, err
			}

			return nil, errors.InternalErr
		}

		return resp, nil
	})
}
