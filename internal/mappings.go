package internal

import (
	"context"
	"fmt"
	"regexp"

	"github.com/shadiestgoat/bankDataDB/data"
)

// Get name & category for a matcher
func (a *API) MapSpecificTransaction(all []*data.Mapping, desc string, amt float64) (name *string, cat *string) {
	for _, m := range all {
		if m.InpAmt != nil && *m.InpAmt != amt {
			continue
		}
		if m.InpText != nil && !(*regexp.Regexp)(m.InpText).MatchString(desc) {
			continue
		}

		applyMatchResult(&name, m.ResName)
		applyMatchResult(&cat, m.ResCategoryID)

		if name != nil && cat != nil {
			return
		}
	}

	return
}

func applyMatchResult(dst **string, src *string) {
	if *dst == nil && src != nil {
		*dst = src
	}
}

func (a *API) UpdateAnyMatchedMappings(ctx context.Context, m *data.Mapping, authorID string) (int, []error) {
	amtUpdated := 0
	var errors []error

	if m.ResCategoryID != nil {
		amt, err := a.store.UpdateTransCatsUsingMapping(ctx, *m.ResCategoryID, authorID, m.InpAmt, (*regexp.Regexp)(m.InpText))
		amtUpdated += amt
		if err != nil {
			errors = append(errors, err)
		}
	}
	if m.ResName != nil {
		amt, err := a.store.UpdateTransNamesUsingMapping(ctx, *m.ResName, authorID, m.InpAmt, (*regexp.Regexp)(m.InpText))
		amtUpdated += amt
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		a.log(ctx).Errorw("Failed to update some transactions", "errors", errors)
	}

	a.log(ctx).Infow("Ran update using mappings", "updated_count", amtUpdated)

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

	return a.store.NewMapping(ctx, authorID, m)
}
