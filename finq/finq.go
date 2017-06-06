package main

import (
	"fmt"

	"github.com/cznic/ql"
)

func doData() {
	ins := ql.MustCompile("BEGIN TRANSACTION; INSERT INTO t VALUES ($1); COMMIT;")

	db, err := ql.OpenMem()
	if err != nil {
		panic(err)
	}

	ctx := ql.NewRWCtx()
	if _, _, err = db.Run(ctx, `
		BEGIN TRANSACTION;
			CREATE TABLE t (c int);
			INSERT INTO t VALUES (1), (2), (3);
		COMMIT;
	`); err != nil {
		panic(err)
	}

	if _, _, err = db.Execute(ctx, ins, int64(42)); err != nil {
		panic(err)
	}

	id := ctx.LastInsertID
	rs, _, err := db.Run(ctx, `SELECT * FROM t WHERE id() == $1`, id)
	if err != nil {
		panic(err)
	}

	if err = rs[0].Do(false, func(data []interface{}) (more bool, err error) {
		fmt.Println(data)
		return true, nil
	}); err != nil {
		panic(err)
	}
}
