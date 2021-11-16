package entity

import (
	"context"

	"github.com/jackc/pgx/v4"
)

func queryRowFuncNoOp(row pgx.QueryFuncRow) error { return nil }

func Query(ctx context.Context, tx pgx.Tx, sql string, args []interface{}, scans []interface{}) error {
	_, err := tx.QueryFunc(ctx, sql, args, scans, queryRowFuncNoOp)
	return err
}
