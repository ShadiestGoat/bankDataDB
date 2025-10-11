package internal_test

import (
	"regexp"
	"testing"

	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMapSpecificTransaction(t *testing.T) {
	t.Run("matching", func(t *testing.T) {
		t.Run("amount", func(t *testing.T) {
			api, _ := utils.NewMockAPI(t)

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
			require.Equal(t, "Name 2", *n)
		})

		t.Run("description", func(t *testing.T) {
			api, _ := utils.NewMockAPI(t)

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
			require.Equal(t, "Name 2", *n)
		})
	})

	t.Run("partial", func(t *testing.T) {
		// Should return different name/category, if a matcher only does 1 thing
		api, _ := utils.NewMockAPI(t)

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
			assert.Equal(t, "Name", *n)
		}
		if assert.NotNil(t, c) {
			assert.Equal(t, "Cat", *c)
		}
	})
}

func TestUpdateAnyMatchedMappings(t *testing.T) {
	api, s := utils.NewMockAPI(t)
	s.EXPECT().UpdateTransCatsUsingMapping(mock.Anything, "cat_id", mock.Anything, mock.Anything, mock.Anything).Return(1, nil)
	s.EXPECT().UpdateTransNamesUsingMapping(mock.Anything, "name", mock.Anything, mock.Anything, mock.Anything).Return(1, nil)

	amt, err := api.UpdateAnyMatchedMappings(t.Context(), &data.Mapping{
		ResName:       utils.Ptr("name"),
		ResCategoryID: utils.Ptr("cat_id"),
	}, "123")

	assert.Empty(t, err)
	assert.Equal(t, 2, amt)
}
