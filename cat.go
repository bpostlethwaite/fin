package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/antzucaro/matchr"
)

var (
	withAddr   *regexp.Regexp = regexp.MustCompile("(.*)\\s+(#.*)")
	withDigits *regexp.Regexp = regexp.MustCompile("(.*)\\s+([0-9]+(-?[0-9]+)+\\s+.*)")
	cities     [10]string     = [10]string{
		"MONTREAL", "MAGOG", "VANCOUVER", "TORONTO", "BOLTON", "CALGARY",
		"MISSISSAUGA", "CANMORE", "OUTREMONT", "CHAMBLY",
	}
)

type Match struct {
	Name string
	Desc string
	dist int
}

func extractName(n string) *Match {
	matches := withAddr.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{Name: matches[1], Desc: matches[2]}
	}

	matches = withDigits.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{Name: matches[1], Desc: matches[2]}
	}

	for _, c := range cities {
		withCity := regexp.MustCompile(fmt.Sprintf("(.*)\\s+(%s.*)", c))
		matches := withCity.FindStringSubmatch(n)
		if len(matches) > 2 {
			return &Match{Name: matches[1], Desc: matches[2]}
		}
	}

	return nil
}

func GetCategory(cat string, cats []string) (string, error) {
	lcat := strings.ToLower(cat)
	for _, c := range cats {
		if strings.ToLower(c) == lcat {
			return c, nil
		}
	}

	// We didn't find a category lets recommend one to the user by
	// lexicographic distance
	ms := make([]Match, len(cats))
	for i, c := range cats {
		ms[i] = Match{
			dist: matchr.DamerauLevenshtein(strings.ToLower(c), lcat),
			Name: c,
		}
	}

	sort.Slice(ms, func(i, j int) bool { return ms[i].dist < ms[j].dist })

	return "", fmt.Errorf("Category %s not found. Did you mean '%s'",
		cat, ms[0].Name)
}

func Categorize(q Query) error {
	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	ftxs, err := QueryTxs(q, txs)
	if err != nil {
		return err
	}

	cats, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	cat, err := GetCategory(q.Cat, cats)
	if err != nil {
		return err
	}

	for i, _ := range ftxs {
		ftxs[i].Category = cat
	}

	txs = AppendDedupeSort(txs, ftxs)
	return store.WriteTransactionTable(txs)
}
