package main

import (
	"fmt"
	"regexp"
	"strings"
)

type Query struct {
	Name string
	Expr string
	Cat  string
}

func filter(txs []Record, f func(Record, int) bool) []Record {
	ftxs := []Record{}
	for i, tx := range txs {
		if f(tx, i) {
			ftxs = append(ftxs, tx)
		}
	}

	return ftxs
}

func QueryTxs(q Query, txs []Record) ([]Record, error) {
	if q.Name != "" {
		name := strings.ToUpper(q.Name)
		return filter(txs, func(t Record, i int) bool {
			return strings.HasPrefix(t.Name, name)
		}), nil

	} else if q.Expr != "" {
		re, err := regexp.Compile(strings.ToUpper(q.Expr))
		if err != nil {
			return nil, err
		}
		return filter(txs, func(t Record, i int) bool {
			return re.MatchString(t.Name)
		}), nil
	} else if q.Cat != "" {
		cat := strings.ToUpper(q.Cat)
		return filter(txs, func(t Record, i int) bool {
			return strings.ToUpper(t.Category) == cat
		}), nil
	}

	return nil, fmt.Errorf("no Query flags provided. Try `fin help query`")
}

func QueryTable(q Query) ([]Record, error) {
	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return nil, err
	}

	return QueryTxs(q, txs)
}
