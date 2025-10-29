package internal_test

import (
	"regexp"
	"testing"

	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/tutils"
	"github.com/shadiestgoat/bankDataDB/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapSpecificTransaction(t *testing.T) {
	t.Run("matching", func(t *testing.T) {
		t.Run("amount", func(t *testing.T) {
			api, _ := tutils.NewMockAPI(t)

			n, _ := api.MapSpecificTransaction(
				[]*data.Mapping{
					{
						InpAmt:  utils.Ptr(1.0),
						ResName: utils.Ptr("Name 1"),
					},
					{
						InpAmt:  utils.Ptr(2.0),
						ResName: utils.Ptr("Name 2"),
					},
				},
				"some desc",
				2.0,
			)

			require.NotNil(t, n)
			require.Equal(t, "Name 2", n.Res)
		})

		t.Run("description", func(t *testing.T) {
			api, _ := tutils.NewMockAPI(t)

			n, _ := api.MapSpecificTransaction(
				[]*data.Mapping{
					{
						InpText: (*data.MarshallableRegexp)(regexp.MustCompilePOSIX(`doesn't match`)),
						ResName: utils.Ptr("Name 1"),
					},
					{
						InpText: (*data.MarshallableRegexp)(regexp.MustCompilePOSIX(`[abs]ome(thing)?`)),
						ResName: utils.Ptr("Name 2"),
					},
				},
				"some desc",
				2.0,
			)

			require.NotNil(t, n)
			require.Equal(t, "Name 2", n.Res)
		})
	})

	t.Run("partial", func(t *testing.T) {
		// Should return different name/category, if a matcher only does 1 thing
		api, _ := tutils.NewMockAPI(t)

		n, c := api.MapSpecificTransaction(
			[]*data.Mapping{
				{
					InpAmt:  utils.Ptr(1.0),
					ResName: utils.Ptr("Name"),
				},
				{
					InpAmt:        utils.Ptr(1.0),
					ResCategoryID: utils.Ptr("Cat"),
					ResName:       utils.Ptr("Some Other Name!"),
				},
			},
			"some desc",
			1.0,
		)

		if assert.NotNil(t, n) {
			assert.Equal(t, "Name", n.Res)
		}
		if assert.NotNil(t, c) {
			assert.Equal(t, "Cat", c.Res)
		}
	})
}
