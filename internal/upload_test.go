package internal_test

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/shadiestgoat/bankDataDB/data"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func findTransactionRowByDesc(rows [][]any, desc string) []any {
	for _, row := range rows {
		if strings.TrimSpace(row[4].(string)) == desc {
			return row
		}
	}

	return nil
}

// Assert that row with desc [desc] exists. Returns nil if it does not exist
func assertTransByDesc(t *testing.T, rows [][]any, desc string) []any {
	trans := findTransactionRowByDesc(rows, desc)

	assert.NotNil(t, trans, "Transaction " + desc + " should exist")

	return trans
}

func date(d string) time.Time {
	t, err := time.Parse("02-01-2006", d)
	if err != nil {
		panic(err)
	}
	return t
}

func TestParseTSV(t *testing.T) {
	t.Run("happy", func(t *testing.T) {
		// trans 1 - exists
		// trans 2 - mapped using desc
		// trans 3 - mapped using desc & amount
		tsv := strings.TrimSpace(`
View current operations and balances - DATE HERE

Account 	ACCOUNT_ID - EUR - CAIXA ACCOUNT
Start Date 	SOME DATE
End Date 	SOME DATE

Op. Date 	Value Date 	Description 	Debit 	Credit 	Balance Accounting 	Balance available 	Categoria (EN) 	
10-08-2025	10-08-2025	ABC 	1.29		15,419.44	--	Diversos 	
10-08-2025	10-08-2025	DEF 	10.79		15,420.73	--	Diversos 	
08-08-2025	06-08-2025	Ghi 	42.17		15,431.52	--	Diversos 	
07-08-2025	06-08-2025	Jkl 	0.99		15,473.69	--	Diversos 	
07-08-2025	05-08-2025	MNO 	3.52		15,474.68	--	Diversos 	
06-08-2025	06-08-2025	PQR 	1,400.00		15,478.20	--	Diversos 	
06-08-2025	06-08-2025	STU 		700.00	16,878.20	---	Diversos 	
06-08-2025	05-08-2025	VXY 	4.90		16,178.20	---	Diversos 	
06-08-2025	06-08-2025	ZAB 	35.49		16,183.10	---	Diversos 	
 	 	 	 	Balance Accounting 	15,419.44 EUR 	 	 	
`)
		api, s := utils.NewMockAPI(t)

		s.EXPECT().GetMappingsForAuthor(mock.Anything, USER_ID).Return([]*data.Mapping{
			{
				InpText: (*data.MarshallableRegexp)(regexp.MustCompilePOSIX("^PQR$")),
				ResName: utils.Ptr("The PQR Transaction"),
			},
			{
				InpText: (*data.MarshallableRegexp)(regexp.MustCompilePOSIX("X")),
				ResName: utils.Ptr("The VXY Transaction"),
			},
			{
				InpAmt: utils.Ptr(700.0),
				ResCategoryID: utils.Ptr("catID"),
			},
		}, nil)

		// Pretend a transaction exists & the rest don't
		s.EXPECT().DoesTransactionExist(
			mock.Anything,
			USER_ID,
			date("05-08-2025"), date("07-08-2025"),
			"MNO", -3.52,
		).Return(true, nil)
		s.EXPECT().DoesTransactionExist(
			mock.Anything,
			mock.Anything,
			mock.Anything, mock.Anything,
			mock.Anything, mock.Anything,
		).Return(false, nil)

		s.EXPECT().InsertCheckpoint(mock.Anything, date("10-08-2025"), 15_419.44)
		s.EXPECT().InsertCheckpoint(mock.Anything, date("08-08-2025"), 15_431.52)
		s.EXPECT().InsertCheckpoint(mock.Anything, date("07-08-2025"), 15_473.69)
		s.EXPECT().InsertCheckpoint(mock.Anything, date("06-08-2025"), 15_478.20)

		// Expect checkpoints to be sent out
		s.EXPECT().SendBatch(mock.Anything, mock.Anything).Return(nil)

		transactionRows := [][]any{}
		s.EXPECT().InsertTransactions(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, v *store.TransactionBatch) (int64, error) {
			transactionRows = v.Rows
			return int64(len(v.Rows)), nil
		})

		resp, err := api.ParseTSV(t.Context(), []byte(tsv), USER_ID)
		require.NoError(t, err)

		assert.Equal(t, 8, resp.NewTransactions)
		assert.Equal(t, 1, resp.SkippedTransactions)
		assert.Equal(t, 7, resp.UnmappedTransactions)

		// already exists - don't enter it again
		assert.Nil(t, findTransactionRowByDesc(transactionRows, "MNO"))

		// Full match (w/ padding)
		if trans := assertTransByDesc(t, transactionRows, "PQR"); trans != nil {
			if assert.NotNil(t, trans[6], "PQR must have a resolved name") {
				assert.Equal(t, *trans[6].(*string), "The PQR Transaction")
			}
		}
		// Partial name match
		if trans := assertTransByDesc(t, transactionRows, "VXY"); trans != nil {
			if assert.NotNil(t, trans[6], "VXY must have a resolved name") {
				assert.Equal(t, *trans[6].(*string), "The VXY Transaction")
			}
		}
		// amount match
		if trans := assertTransByDesc(t, transactionRows, "STU"); trans != nil {
			if assert.NotNil(t, trans[7], "STU must have a resolved category") {
				assert.Equal(t, *trans[7].(*string), "catID")
			}
		}
	})
}
