package external

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
	"github.com/shadiestgoat/bankDataDB/log"
)

func Router(api *internal.API, store store.Store) chi.Router {
	r := chi.NewRouter()

	r.Use(
		makeMiddleware(api, middlewareLog),
		middleware.AllowContentType("application/json"),
		middleware.CleanPath,
		middleware.Compress(5),
		middleware.Recoverer,
	)

	mountUser(api, r)
	r.Group(func(r chi.Router) {
		r.Use(makeMiddleware(api, middlewareAuthUser))

		r.Route(`/transactions`, func(r chi.Router) { mountTransactions(api, r) })
		r.Route(`/upload`, func(r chi.Router) { mountUpload(api, r) })
		r.Route(`/mappings`, func(r chi.Router) { routeMappings(r, api, store) })
		r.Route(`/categories`, func(r chi.Router) { routeCategories(r, api, store) })
	})

	return r
}

func logRoute(r chi.Route, indent int) {
	for m := range r.Handlers {
		fmt.Println(strings.Repeat(" ", indent * 2) + m, r.Pattern)
	}

	if r.SubRoutes == nil {
		return
	}

	for _, s := range r.SubRoutes.Routes() {
		logRoute(s, indent + 1)
	}
}

func handleHTTPError(a *internal.API, w http.ResponseWriter, ctx context.Context, err error) {
	l := a.Logger()(ctx)
	if err, ok := err.(errors.GenericHTTPError); ok {
		err.Render(l, w)
	} else {
		l.Errorw("Unknown error returned by an endpoint", "error", err)
		errors.InternalErr.Render(l, w)
	}
}

func makeMiddleware(a *internal.API, mw func(a *internal.API, r *http.Request) (*http.Request, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			newReq, err := mw(a, r)
			if newReq != nil {
				r = newReq
			}

			if err != nil {
				handleHTTPError(a, w, r.Context(), err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func middlewareLog(a *internal.API, r *http.Request) (*http.Request, error) {
	ctx := log.ContextSet(r.Context(), a.Logger()(r.Context()), "method", r.Method, "route", r.Pattern)

	return r.WithContext(ctx), nil
}

func defHTTP(r chi.Router, m string, p string, a *internal.API, h func(r *http.Request) (any, errors.GenericHTTPError)) {
	r.Method(m, p, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp, httpErr := h(r)
		if httpErr != nil {
			handleHTTPError(a, w, r.Context(), httpErr)
			return
		}
		if resp == nil {
			w.WriteHeader(204)
			return
		}

		enc, err := json.Marshal(resp)
		if err != nil {
			a.Logger()(r.Context()).Errorw("Failed to marshal response for an endpoint", "error", err)
			handleHTTPError(a, w, r.Context(), errors.InternalErr)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(enc)
	}))
}

func defHTTPRead[T any](r chi.Router, m string, p string, a *internal.API, h func(r *http.Request, b T) (any, errors.GenericHTTPError)) {
	defHTTP(r, m, p, a, func(r *http.Request) (any, errors.GenericHTTPError) {
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		a.Logger()(r.Context()).Debugf("Hit")

		var s T
		err := dec.Decode(&s)
		if err != nil {
			a.Logger()(r.Context()).Debugf("Slayed", "err", err)
			return nil, errors.BadInput
		}

		return h(r, s)
	})
}

type RouteContextKey int

const (
	CTX_USER_ID RouteContextKey = iota
)

type RespPages struct {
	Total int `json:"total"`
	Data  any `json:"data"`
}

// handler needs to return:
// 1. Current data
// 2. The total count
func defHTTPPage(r chi.Router, a *internal.API, allowedCols []string, h func(r *http.Request, size, off int, orderBy string, asc bool) (any, int, errors.GenericHTTPError)) {
	defHTTP(r, `GET`, `/`, a, func(r *http.Request) (any, errors.GenericHTTPError) {
		page, size, orderBy, asc := 0, 50, allowedCols[0], false
		q := r.URL.Query()

		if inp := q.Get("page"); inp != "" {
			v, err := strconv.ParseUint(inp, 10, 16)
			if err != nil {
				return nil, errors.BadInput
			}
			page = int(v)
		}
		if inp := q.Get("size"); inp != "" {
			v, err := strconv.ParseUint(inp, 10, 16)
			if err != nil {
				return nil, errors.BadInput
			}
			size = int(v)
		}
		if inp := q.Get("order"); inp != "" {
			if !slices.Contains(allowedCols, inp) {
				return nil, errors.BadInput
			}
			orderBy = inp
		}
		if inp := q.Get("asc"); inp != "" {
			if !slices.Contains([]string{"true", "false"}, inp) {
				return nil, errors.BadInput
			}
			asc = inp == "true"
		}

		data, total, err := h(r, size, page*size, orderBy, asc)
		if err != nil {
			return nil, err
		}

		return &RespPages{
			Total: total,
			Data:  data,
		}, err
	})
}

type RespCreated struct {
	ID string `json:"id"`
}

type RespDeleted struct {
	DidDelete bool `json:"deleted"`
}
