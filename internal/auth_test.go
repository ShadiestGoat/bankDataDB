package internal_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/internal"
	"github.com/shadiestgoat/bankDataDB/tutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	USER_NAME = "name"
	USER_PASS = "pass"
)

func TestLogin(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {

		t.Run("username", func(t *testing.T) {
			api, s := tutils.NewMockAPI(t)
			badName := USER_NAME + "123"
			s.EXPECT().GetUserByName(mock.Anything, badName).Return(nil, pgx.ErrNoRows)

			str, err := api.Login(t.Context(), badName, USER_PASS)
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, errors.BadAuth)
			}
			assert.Empty(t, str)
		})

		t.Run("password", func(t *testing.T) {
			api, s := tutils.NewMockAPI(t)
			s.EXPECT().GetUserByName(mock.Anything, USER_NAME).Return(
				&store.User{Password: "Something thats CLEARLY not hashed"},
				nil,
			)

			str, err := api.Login(t.Context(), USER_NAME, USER_PASS+"123")
			if assert.Error(t, err) {
				assert.ErrorIs(t, err, errors.BadAuth)
			}
			assert.Empty(t, str)
		})
	})

	t.Run("valid", func(t *testing.T) {
		hashed, err := internal.UtilPasswordGen(USER_PASS)
		require.NoError(t, err)

		api, s := tutils.NewMockAPI(t)
		s.EXPECT().GetUserByName(mock.Anything, USER_NAME).Return(
			&store.User{Password: string(hashed)},
			nil,
		)

		str, err := api.Login(t.Context(), USER_NAME, USER_PASS)
		assert.NoError(t, err)
		require.NotEmpty(t, str)

		tok, _, err := jwt.NewParser().ParseUnverified(str, jwt.MapClaims{})
		require.NoError(t, err)
		_, err = tok.Claims.GetIssuedAt()
		assert.NoError(t, err)
		_, err = tok.Claims.GetExpirationTime()
		assert.NoError(t, err)
	})
}

func TestExchangeToken(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		api, s := tutils.NewMockAPI(t)

		tok, err := api.NewToken(t.Context(), USER_ID)
		require.NoError(t, err)

		s.EXPECT().GetUserUpdatedAt(mock.Anything, USER_ID).Return(
			time.Now().Add(-12*time.Hour),
			nil,
		)

		id := api.ExchangeToken(t.Context(), tok)
		require.NotNil(t, id)
		require.Equal(t, USER_ID, *id)
	})

	t.Run("before_updated_at", func(t *testing.T) {
		api, s := tutils.NewMockAPI(t)

		tok, err := api.NewToken(t.Context(), USER_ID)
		require.NoError(t, err)

		s.EXPECT().GetUserUpdatedAt(mock.Anything, USER_ID).Return(
			time.Now().Add(12*time.Hour),
			nil,
		)

		id := api.ExchangeToken(t.Context(), tok)
		require.Nil(t, id)
	})
}
