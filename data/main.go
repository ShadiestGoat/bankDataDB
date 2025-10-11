package data

import (
	"encoding/json"
	"regexp"
	"time"
)

type MarshallableRegexp regexp.Regexp

type Mapping struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	InpText *MarshallableRegexp `json:"inputText,omitempty"`
	InpAmt  *float64            `json:"inputAmount,omitempty"`

	ResName       *string `json:"resName,omitempty"`
	ResCategoryID *string `json:"resCategoryID,omitempty"`

	Priority int `json:"priority"`
}

func (m *MarshallableRegexp) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	regexp, err := regexp.CompilePOSIX(str)
	if err != nil {
		return err
	}

	*m = MarshallableRegexp(*regexp)

	return nil
}

func (m *MarshallableRegexp) MarshalJSON() ([]byte, error) {
	return json.Marshal((*regexp.Regexp)(m).String())
}

type Transactions struct {
	SettledAt time.Time `json:"settledAt"`
	AuthedAt time.Time `json:"authedAt"`
	Desc string `json:"description"`
	Amount string `json:"amount"`
	ResolvedName *string `json:"resolvedName,omitempty"`
	ResolvedCategoryID *string `json:"resolvedCategoryId,omitempty"`
}
