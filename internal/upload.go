package internal

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/shadiestgoat/bankDataDB/db/store"
	"github.com/shadiestgoat/bankDataDB/external/errors"
	"github.com/shadiestgoat/bankDataDB/log"
)

func parseNum(l log.Logger, s string) (float64, error) {
	if s == "" {
		return 0, nil
	}

	f, err := strconv.ParseFloat(strings.ReplaceAll(s, ",", ""), 64)
	if err != nil {
		l.Warnf("Can't parse number (%s): %v", s, err)
		return f, err
	}

	return f, nil
}

type InsertResp struct {
	NewTransactions      int `json:"newTransactions"`
	SkippedTransactions  int `json:"skippedTransactions"`
	UnmappedTransactions int `json:"unmappedTransactions"`
}

func (a *API) ParseTSV(ctx context.Context, tsv []byte, authorID string) (*InsertResp, error) {
	sec := strings.Split(string(tsv), "\n\n")
	if len(sec) != 3 {
		return nil, errors.BadTSV
	}

	lines := strings.Split(sec[2], "\n")
	if len(lines) < 2 {
		return nil, errors.BadTSV
	}

	mappings, err := a.store.GetMappingsForAuthor(ctx, authorID)
	if err != nil {
		return nil, err
	}

	resp := &InsertResp{
		NewTransactions:      0,
		SkippedTransactions:  0,
		UnmappedTransactions: 0,
	}

	var lastRowCheckpointDate time.Time
	batchCheckpoints := &pgx.Batch{}
	batchTrans := &store.TransactionBatch{}

	for _, l := range lines[1 : len(lines)-1] {
		// Op. Date 	Value Date 	Description 	Debit 	Credit 	Balance Accounting 	Balance available 	Categoria (EN)
		cols := strings.Split(l, "\t")
		if len(cols) < 8 {
			fmt.Println("(!)", len(cols))
			a.log(ctx).Warnf("Wrong # of columns (expected 8, got %d)", len(cols))
			continue
		}

		authedAt, err := time.Parse("02-01-2006", cols[1])
		if err != nil {
			a.log(ctx).Warnf("Can't parse date (%v): %v", cols[1], err)
			continue
		}

		settledAt, err := time.Parse("02-01-2006", cols[0])
		if err != nil {
			a.log(ctx).Warnf("Can't parse date (%v): %v", cols[0], err)
			continue
		}

		desc := strings.TrimSpace(cols[2])
		deb, err := parseNum(a.log(ctx), cols[3])
		if err != nil {
			continue
		}
		cred, err := parseNum(a.log(ctx), cols[4])
		if err != nil {
			continue
		}
		amt := cred - deb

		amtAfter, err := parseNum(a.log(ctx), cols[5])
		if err != nil {
			continue
		}

		ctx := log.ContextSet(ctx, a.log(ctx), "desc", desc, "amt", amt, "authed_at", authedAt)

		exist, err := a.store.DoesTransactionExist(ctx, authorID, authedAt, settledAt, desc, amt)
		fmt.Println(authorID, authedAt, settledAt, desc, amt, exist, err)

		if err != nil {
			a.log(ctx).Errorf("Can't verify transaction existing: %v", err)
			continue
		}
		if exist {
			a.log(ctx).Infof("Skipping transaction insert because it already exists")
			resp.SkippedTransactions++
			continue
		}

		resolvedName, resolvedCat := a.MapSpecificTransaction(mappings, desc, amt)
		if resolvedCat == nil && resolvedName == nil {
			resp.UnmappedTransactions++
		}

		batchTrans.Insert(authedAt, settledAt, authorID, desc, amt, resolvedName, resolvedCat)

		if lastRowCheckpointDate != settledAt {
			a.store.InsertCheckpoint(batchCheckpoints, settledAt, amtAfter)
		}
		lastRowCheckpointDate = settledAt
	}

	a.log(ctx).Infow("Writing transactions to db", "amount", len(batchTrans.Rows))
	c, err := a.store.InsertTransactions(ctx, batchTrans)
	if err != nil {
		return nil, err
	}
	resp.NewTransactions = int(c)

	err = a.store.SendBatch(ctx, batchCheckpoints)
	if err != nil {
		// Not a hard stopping err
		a.log(ctx).Errorw("Couldn't insert checkpoints", "error", err)
	}

	return resp, nil
}
