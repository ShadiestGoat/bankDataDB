package internal

import (
	"context"
	"math/rand"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/snownode"
	"golang.org/x/crypto/bcrypt"
)

const (
	JWT_ISSUER     = "bank_author"
	TOKEN_DURATION = 2 * time.Hour
)

type JWTConfig struct {
	Secret []byte
}

var (
	jwtParser = jwt.NewParser(
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(JWT_ISSUER),
		jwt.WithIssuedAt(),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS512.Name}),
	)
)

func (a *API) ExchangeToken(ctx context.Context, t string) *string {
	tok, err := jwtParser.Parse(t, func(t *jwt.Token) (any, error) {
		return a.cfg.JWT.Secret, nil
	})
	if err != nil {
		return nil
	}
	c, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		a.log(ctx).Errorw("Received a token w/ claims that aren't a MapClaims???")
		return nil
	}

	rawUserID, ok := c["usr"]
	if !ok {
		return nil
	}
	userID, ok := rawUserID.(string)
	if !ok {
		return nil
	}

	issuedAt, err := c.GetIssuedAt()
	if err != nil {
		return nil
	}

	userUpdatedAt, err := a.store.GetUserUpdatedAt(ctx, userID)
	if err != nil || userUpdatedAt.After(issuedAt.Time) {
		return nil
	}

	return &userID
}

func (a *API) Login(ctx context.Context, username, inpPassword string) (string, error) {
	usr, err := a.store.GetUserByName(ctx, username)
	if err != nil {
		a.log(ctx).Debugw("User doesn't exist", "name", username)
		time.Sleep(time.Duration(rand.Float64()) * time.Second)
		return "", errors.BadAuth
	}

	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(inpPassword)); err != nil {
		a.log(ctx).Debugw("User has a bad password", "name", username)
		time.Sleep(time.Duration(rand.Float64()) * time.Second)
		return "", errors.BadAuth
	}

	a.log(ctx).Debugw("Logged in", "username", username)

	return a.NewToken(ctx, usr.ID)
}

func UtilPasswordGen(pass string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(pass), 0)
}

// Creates a user in the DB. If the ID is empty, it will be auto-created
func (a *API) CreateUser(ctx context.Context, id, name, password string) error {
	if id == "" {
		id = snownode.NewID()
	}

	pass, err := UtilPasswordGen(password)
	if err != nil {
		return err
	}

	_, err = a.db.Exec(ctx, `INSERT INTO users (id, username, password) VALUES ($1, $2, $3)`, id, name, pass)

	return err
}

func (a *API) NewToken(ctx context.Context, userID string) (string, error) {
	str, err := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"iss": JWT_ISSUER,
		"iat": float64(time.Now().Unix()),
		"exp": float64(time.Now().Add(TOKEN_DURATION).Unix()),
		"usr": userID,
	}).SignedString(a.cfg.JWT.Secret)
	if err != nil {
		a.log(ctx).Errorw("Failed to create a JWT token", "error", err)
	}

	return str, err
}
