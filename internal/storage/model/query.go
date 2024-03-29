package model

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func queryRowFuncNoOp(pgx.QueryFuncRow) error { return nil }

func query(ctx context.Context, tx pgx.Tx, sql string, args []interface{}, scans []interface{}) error {
	_, err := tx.QueryFunc(ctx, sql, args, scans, queryRowFuncNoOp)
	return err
}

func queryUpdateDelete(ctx context.Context, tx pgx.Tx, sql string, args []interface{}) (bool, error) {
	tag, err := tx.QueryFunc(ctx, sql, args, nil, queryRowFuncNoOp)
	return tag.RowsAffected() > 0, err
}
