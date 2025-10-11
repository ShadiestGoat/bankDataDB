package internal

import (
	"context"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"
)

type SavableCategory struct {
	Color string `json:"color"`
	Icon  string `json:"icon"`
	Name  string `json:"name"`
}

type ValidationErr struct {
	Details []string
}

func (v ValidationErr) Error() string {
	return "Failed to validate: " + strings.Join(v.Details, ", ")
}

func (s SavableCategory) Validate() error {
	e := []string{}
	if uniseg.GraphemeClusterCount(s.Icon) != 1 {
		e = append(e, "icon: Needs to only be 1 character")
	}
	if len(s.Color) != 6 {
		e = append(e, "color: Color must have 6 characters! So, hex color")
	} else {
		_, err := strconv.ParseUint(s.Color, 16, 0)
		if err != nil {
			e = append(e, "color: Color is not right!")
		}
	}

	if len(e) != 0 {
		return ValidationErr{e}
	}

	return nil
}

func (a *API) CreateCategory(ctx context.Context, c *SavableCategory, authorID string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	return a.store.CreateCategory(ctx, authorID, c.Name, c.Icon, c.Color)
}