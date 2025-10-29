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

func (a *API) MappingCreate(ctx context.Context, authorID string, m *data.Mapping, retroactivelyMap bool) (string, int, error) {
	if m.ID != "" {
		return "", 0, fmt.Errorf("id present")
	}
	if err := a.validateMapping(ctx, authorID, m); err != nil {
		return "", 0, err
	}

	var (
		id string
		affectedTrans int
	)

	err := a.store.TxFunc(ctx, func(s store.Store) error {
		mappingID, err := s.MappingInsert(ctx, authorID, m)
		if err != nil {
			return err
		}
	
		if !retroactivelyMap {
			return nil
		}
	
		id = mappingID
		m.ID = id
	
		if m.ResCategoryID != nil {
			affected, err := s.TransMapsMapExisting(ctx, false, authorID, m)
			if err != nil {
				return err
			}
	
			affectedTrans += affected
		}
		if m.ResName != nil {
			affected, err := s.TransMapsMapExisting(ctx, true, authorID, m)
			if err != nil {
				return err
			}
	
			affectedTrans += affected
		}

		return nil
	})

	if err != nil {
		return "", 0, err
	}

	return id, affectedTrans, nil
}

func cmpPtr[T comparable](a, b *T) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	return *a == *b
}

func (a *API) MappingUpdate(ctx context.Context, authorID string, oldMapping, newMapping *data.Mapping, retroactive bool) (error) {
	newMapping.ID = oldMapping.ID

	if err := a.validateMapping(ctx, authorID, newMapping); err != nil {
		return err
	}

	return a.store.TxFunc(ctx, func(s store.Store) error {
		err := s.MappingReset(
			ctx,
			&store.MappingResetParams{
				ID:          newMapping.ID,
				Name:        newMapping.Name,
				Priority:    int32(newMapping.Priority),
				TransText:   newMapping.InpText.TextNil(),
				TransAmount: newMapping.InpAmt,
				ResName:     newMapping.ResName,
				ResCategory: newMapping.ResCategoryID,
			},
		)
		if err != nil {
			return err
		}

		matchersChanged := !cmpPtr(oldMapping.InpAmt, newMapping.InpAmt) ||
						   !cmpPtr(oldMapping.InpText.TextNil(), newMapping.InpText.TextNil()) ||
						   oldMapping.Priority != newMapping.Priority
		remapNames := cmpPtr(oldMapping.ResName, newMapping.ResName)
		remapCats := cmpPtr(oldMapping.ResCategoryID, newMapping.ResCategoryID)

		if !retroactive {
			if matchersChanged || (remapCats && remapNames) {
				return s.TransMapsOrphanAll(ctx, newMapping.ID)
			} else if remapNames {
				return s.TransMapsOrphanNames(ctx, newMapping.ID)
			}  else if remapCats {
				return s.TransMapsOrphanCategories(ctx, newMapping.ID)
			}

			return nil
		}

		// if matchers changed, then to retroactively be all good, we gotta re-map everything
		if matchersChanged {
			if err := s.TransMapsCleanAll(ctx, newMapping.ID); err != nil {
				return err
			}
			if _, err := s.TransMapsMapExisting(ctx, true, authorID, newMapping); err != nil {
				return err
			}
			if _, err := s.TransMapsMapExisting(ctx, false, authorID, newMapping); err != nil {
				return err
			}

			return nil
		}

		// If only a result changed, then like we can just update the already matched transactions and be good with it data against the res
		if remapNames {
			if err := s.TransMapsUpdateLinkedNames(ctx, newMapping.ID, newMapping.ResName); err != nil {
				return err
			}
		}
		if remapCats {
			if err := s.TransMapsUpdateLinkedNames(ctx, newMapping.ID, newMapping.ResCategoryID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (a *API) MappingDelete(ctx context.Context, mappingID string, retroactive bool) error {
	return a.store.TxFunc(ctx, func(s store.Store) error {
		if retroactive {
			if err := a.store.TransMapsCleanAll(ctx, mappingID); err != nil {
				return err
			}
		} else {
			if err := a.store.TransMapsOrphanAll(ctx, mappingID); err != nil {
				return err
			}
		}

		return a.store.MappingDelete(ctx, mappingID)
	})
}
