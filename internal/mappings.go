package internal

import (
	"context"
	"fmt"
	"regexp"

	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db/store"
)

type MappingRes struct {
	Res string
	MappingID string
}

func (m *MappingRes) SafeValue() *string {
	if m == nil {
		return nil
	}
	return &m.Res
}

// Get name & category for a matcher
func (a *API) MapSpecificTransaction(all []*data.Mapping, desc string, amt float64) (name *MappingRes, cat *MappingRes) {
	for _, m := range all {
		if m.InpAmt != nil && *m.InpAmt != amt {
			continue
		}
		if m.InpText != nil && !(*regexp.Regexp)(m.InpText).MatchString(desc) {
			continue
		}

		applyMatchResult(&name, m.ResName, m.ID)
		applyMatchResult(&cat, m.ResCategoryID, m.ID)

		if name != nil && cat != nil {
			return
		}
	}

	return
}

func applyMatchResult(dst **MappingRes, src *string, mappingID string) {
	if *dst == nil && src != nil {
		*dst = &MappingRes{
			Res:       *src,
			MappingID: mappingID,
		}
	}
}

func remapType(ctx context.Context, s store.Store, m *data.Mapping, newVal *string, authorID string, forName bool) (int, error) {
	var err error
	if forName {
		err = s.TransMapsRmNames(ctx, m.ID)
	} else {
		err = s.TransMapsRmCategories(ctx, m.ID)
	}

	if err != nil || newVal == nil {
		return 0, err
	}

	iter, amt, err := s.TransMapsMapExisting(ctx, forName, *newVal, authorID, m.InpAmt, (*regexp.Regexp)(m.InpText))
	if err != nil {
		return 0, err
	}
	defer iter.SafeClose()

	err = s.TransMapsInsert(ctx, *iter, m.ID, forName)
	if err != nil {
		return 0, err
	}

	return amt, nil
}

func (a *API) TransRemapForOneMapping(ctx context.Context, m *data.Mapping, remapNames, remapCategories bool, authorID string) (int, []error) {
	amtUpdated := 0
	var errors []error

	a.store.TxFunc(ctx, func(s store.Store) error {
		if remapCategories {
			amt, err := remapType(ctx, s, m, m.ResCategoryID, authorID, false)
			if err != nil {
				return err
			}
			amtUpdated += amt
		}
		if remapNames {
			amt, err := remapType(ctx, s, m, m.ResName, authorID, true)
			if err != nil {
				return err
			}
			amtUpdated += amt
		}

		return nil
	})

	return amtUpdated, errors
}

func (a *API) validateMapping(ctx context.Context, authorID string, inp *data.Mapping) error {
	e := []string{}

	if inp.Name == "" {
		e = append(e, "name: required")
	}
	if inp.InpText == nil && inp.InpAmt == nil {
		e = append(
			e,
			"inpDescRegex: at least 1 selector must be defined",
			"inpAmount: at least 1 selector must be defined",
		)
	}
	if inp.ResCategoryID == nil && inp.ResName == nil {
		e = append(
			e,
			"resName: at least 1 result must be defined",
			"resCategory: at least 1 result must be defined",
		)
	}
	if len(e) != 0 && inp.ResCategoryID != nil {
		if ok, err := a.store.DoesCategoryExist(ctx, authorID, *inp.ResCategoryID); err == nil && !ok {
			e = append(e, "resCategory: Does not exist")
		}
	}

	if len(e) != 0 {
		return &ValidationErr{e}
	}

	return nil
}

func (a *API) CreateMapping(ctx context.Context, authorID string, m *data.Mapping) (string, error) {
	if m.ID != "" {
		return "", fmt.Errorf("id present")
	}
	if err := a.validateMapping(ctx, authorID, m); err != nil {
		return "", err
	}

	return a.store.MappingInsert(ctx, authorID, m)
}
