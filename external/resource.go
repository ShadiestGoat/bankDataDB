package external

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

const MAX_RESOURCES = 50

type Resource[T internal.ResourceValue] struct {
	Path  string
	Table string

	// Verify a sample & return an array of errors
	Verify func(sample T) []string
	// Callback when new was created
	OnNew func(sample T, authorID string)
}

type ResourceRespID struct {
	ID string `json:"id,omitempty"`
}

type RespMsg struct {
	Message string `json:"message"`
}

func resourcePaths[T internal.ResourceValue](api *internal.API, res *Resource[T], r chi.Router) {
	r.Route("/"+res.Path, func(r chi.Router) {
		defHTTP(r, `GET`, `/`, api, func(r *http.Request) (any, errors.GenericHTTPError) {
			var s T
			api.Logger()(r.Context()).Debugw("Mewo???")

			resp, err := api.ResourceGetAll(r.Context(), res.Table, getUserID(r), s)
			if err != nil {
				api.Logger()(r.Context()).Errorw("Can't get resource", "error", err, "resource", res.Table)
				return nil, errors.InternalErr
			}

			return resp, nil
		})
		defHTTPRead(r, `POST`, `/`, api, func(r *http.Request, s T) (any, errors.GenericHTTPError) {
			inpErrors := res.Verify(s)
			if len(inpErrors) != 0 {
				return nil, &errors.OnDemandHTTPError{
					Status:  400,
					Message: "Bad user input",
					Details: inpErrors,
				}
			}
			uID := getUserID(r)

			c, err := api.ResourceGetCount(r.Context(), res.Table, uID)
			if err != nil {
				return nil, errors.InternalErr
			}
			if c > MAX_RESOURCES {
				return nil, &errors.OnDemandHTTPError{
					Status:  400,
					Message: "Too many resources already exist!",
				}
			}

			id, err := api.ResourceInsert(r.Context(), res.Table, s, uID)
			if err != nil {
				return nil, errors.InternalErr
			}

			if res.OnNew != nil {
				go res.OnNew(s, uID)
			}

			return &ResourceRespID{id}, nil
		})
		r.Route(`/{id}`, func(r chi.Router) {
			r.Use(makeMiddleware(api, func(a *internal.API, r *http.Request) (*http.Request, error) {
				exist, err := a.ResourceVerifyExists(r.Context(), res.Table, getUserID(r), chi.URLParam(r, `id`))
				if err != nil {
					return nil, errors.InternalErr
				}
				if !exist {
					return nil, errors.NotFound
				}

				return r, nil
			}))

			defHTTP(r, `GET`, `/`, api, func(r *http.Request) (any, errors.GenericHTTPError) {
				var resp T
				err := api.ResourceGetSpecific(r.Context(), res.Table, chi.URLParam(r, `id`), resp)
				if err != nil {
					return nil, errors.InternalErr
				}

				return resp, nil
			})

			defHTTPRead(r, `POST`, `/`, api, func(r *http.Request, s T) (any, errors.GenericHTTPError) {
				inpErrors := res.Verify(s)
				if len(inpErrors) != 0 {
					return nil, &errors.OnDemandHTTPError{
						Status:  400,
						Message: "Bad user input",
						Details: inpErrors,
					}
				}

				uID := chi.URLParam(r, `id`)
				err := api.ResourceUpdate(r.Context(), res.Table, s, uID)
				if err != nil {
					return nil, errors.InternalErr
				}

				if res.OnNew != nil {
					go res.OnNew(s, uID)
				}

				return &RespMsg{"Updated"}, nil
			})

			defHTTP(r, `DELETE`, `/`, api, func(r *http.Request) (any, errors.GenericHTTPError) {
				err := api.ResourceDrop(r.Context(), res.Table, chi.URLParam(r, `id`))
				if err != nil {
					return nil, errors.InternalErr
				}

				return &RespMsg{"Deleted"}, nil
			})
		})
	})
}

func mountResources(api *internal.API, r chi.Router) {
	resourcePaths(api, &Resource[*internal.ResMappings]{
		Path:  "mappings",
		Table: "mappings",
		Verify: func(inp *internal.ResMappings) []string {
			e := []string{}

			if inp.Name == "" {
				e = append(e, "name: required")
			}
			if inp.InpText == nil && inp.InpAmt == nil {
				e = append(e, "inpDescRegex, inpAmount: at least 1 must be present")
			}
			if inp.ResCategoryID == nil && inp.ResName == nil {
				e = append(e, "resName, resCategory: at least 1 must be present")
			}

			return e
		},
		OnNew: func(sample *internal.ResMappings, authorID string) {
			api.UpdateAnyMatchedMappings(context.Background(), &sample.Mapping, authorID)
		},
	}, r)
}
