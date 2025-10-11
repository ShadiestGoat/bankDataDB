package internal

import (
	"context"
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
