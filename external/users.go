package external

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
)

type ReqLogin struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RespLogin struct {
	Token string `json:"token"`
}

func mountUser(api *internal.API, r chi.Router) {
	defHTTPRead(r, `POST`, `/login`, api, func(r *http.Request, b ReqLogin) (any, errors.GenericHTTPError) {
		tok, err := api.Login(r.Context(), b.Username, b.Password)
		if err != nil {
			if err, ok := err.(errors.GenericHTTPError); ok {
				return nil, err
			}
			return nil, errors.InternalErr
		}

		return &RespLogin{tok}, nil
	})
}

func getUserID(r *http.Request) string {
	return r.Context().Value(CTX_USER_ID).(string)
}

func middlewareAuthUser(a *internal.API, r *http.Request) (*http.Request, error) {
	t := r.Header.Get("Authorization")
	fmt.Println("Meow meow??")
	if t == "" {
		return nil, errors.NoAuthProvided
	}

	res := a.ExchangeToken(r.Context(), t)
	if res == nil {
		return nil, errors.BadAuth
	}

	return r.WithContext(context.WithValue(r.Context(), CTX_USER_ID, *res)), nil
}
