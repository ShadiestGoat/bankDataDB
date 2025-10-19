package internal

import (
	"context"
	"strconv"

	"github.com/rivo/uniseg"
)

type SavableCategory struct {
	Color string `json:"color"`
	Icon  string `json:"icon"`
	Name  string `json:"name"`
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

	if len(s.Name) == 0 {
		e = append(e, "name: required")
	}

	if len(e) != 0 {
		return &ValidationErr{e}
	}

	return nil
}

func (a *API) CreateCategory(ctx context.Context, c *SavableCategory, authorID string) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}

	return a.store.NewCategory(ctx, authorID, c.Name, c.Icon, c.Color)
}

func (a *API) UpdateCategory(ctx context.Context, id string, c *SavableCategory) error {
	if err := c.Validate(); err != nil {
		return err
	}

	return a.store.ResetCategoryData(ctx, id, c.Name, c.Color, c.Icon)
}
