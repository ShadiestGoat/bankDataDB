package errors

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/shadiestgoat/bankDataDB/log"
)

//go:generate go run -C ./generator . ../errors.go

type GenericHTTPError interface {
	error
	Render(l log.Logger, w http.ResponseWriter)
}

type HTTPError struct {
	Status int
	rendered []byte
}

func (e HTTPError) Error() string {
	return fmt.Sprintf("%d: %s", e.Status, string(e.rendered))
}

func (h HTTPError) HTTP() bool {
	return true
}

func (e HTTPError) Render(_ log.Logger, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	w.Write(e.rendered)
}

type OnDemandHTTPError struct {
	Status int `json:"-"`
	Message string `json:"error"`
	Details []string `json:"details,omitempty"`
}

func (e OnDemandHTTPError) Error() string {
	return fmt.Sprintf("%d: On Demand w/ %d details: %s", e.Status, len(e.Details), e.Message)
}

func (e OnDemandHTTPError) Render(l log.Logger, w http.ResponseWriter) {
	enc, err := json.Marshal(e)
	if err != nil {
		l.Errorw("Can't marshal an on-demand error", "error", err, "message", e.Message)
		InternalErr.Render(l, w)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	w.Write(enc)
}
