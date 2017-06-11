package main

import (
	"context"
	"regexp"
	"strings"

	"googlemaps.github.io/maps"
)

var replDigits *regexp.Regexp = regexp.MustCompile("(.*\\s+)([0-9]+(-?[0-9]+)+)(\\s+.*)")

func Recommend() ([]Record, error) {
	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return nil, err
	}

	cats, err := store.ReadCategoryTable()
	if err != nil {
		return nil, err
	}

	// First see if there are any recommendations we can make based on
	// internal consistencies.
	internalTxs := InternalSearch(txs)

	// Now check Google places to see if there any other recommendations.
	// Pass in all txs updated with Internal search.
	matches, err := GooglePlaces(AppendDedupeSort(txs, internalTxs), cats)
	if err != nil {
		return nil, err
	}

	placeTxs := []Record{}
	for _, m := range matches {
		if m.Category != "" {
			m.Record.Category = m.Category
			placeTxs = append(placeTxs, m.Record)
		}
	}

	// return only updated transactions
	return append(internalTxs, placeTxs...), nil
}

func PlaceSearch() ([]Match, error) {
	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return nil, err
	}

	cats, err := store.ReadCategoryTable()
	if err != nil {
		return nil, err
	}

	// Now check Google places to see if there any other recommendations.
	// Pass in all txs updated with Internal search.
	matches, err := GooglePlaces(txs, cats)
	if err != nil {
		return nil, err
	}

	unmatched := []Match{}
	for _, m := range matches {
		if m.Category == "" {
			unmatched = append(unmatched, m)
		}
	}

	return unmatched, nil
}

func InternalSearch(txs []Record) []Record {
	uncat := []Record{}
	short := []Record{}
	ftxs := []Record{}

	for _, tx := range txs {
		if tx.Category == UNCATEGORIZED {
			uncat = append(uncat, tx)
		} else {
			if len(tx.Name) >= COMPARE_CHARS {
				tx.Name = tx.Name[:COMPARE_CHARS]
			}

			short = append(short, tx)
		}
	}

	for _, uc := range uncat {
		for _, s := range short {
			if strings.HasPrefix(uc.Name, s.Name) {
				uc.Category = s.Category
				ftxs = append(ftxs, uc)
				break
			}
		}
	}

	return ftxs
}

func PlaceType(query string) (string, error) {
	c, err := maps.NewClient(maps.WithAPIKey(ConfigCreds().GoogleMapsApiKey))
	if err != nil {
		return "", err
	}

	tSearch := maps.TextSearchRequest{
		Query: query,
	}
	resp, err := c.TextSearch(context.Background(), &tSearch)
	if err != nil {
		return "", err
	}

	var category string
	if len(resp.Results) > 0 {
		if len(resp.Results[0].Types) > 0 {
			category = resp.Results[0].Types[0]
		}
	}

	return category, nil
}

func GooglePlaces(txs []Record, cats [][]string) ([]Match, error) {

	matches := []Match{}
	for _, tx := range txs {
		if tx.Category != UNCATEGORIZED {
			continue
		}

		// remove weird digits and whatnot
		subName := replDigits.ReplaceAllString(tx.Name, "$1$4")
		gcat, err := PlaceType(subName)
		if err != nil {
			if strings.Contains(err.Error(), "ZERO_RESULTS") {
				continue
			}
			return nil, err
		}

		cat := getCategoryFromRelated(gcat, cats)
		matches = append(matches, Match{
			Name:      subName,
			PlaceType: gcat,
			Category:  cat,
			Record:    tx,
		})
	}

	return matches, nil
}

func getCategoryFromRelated(rel string, cats [][]string) string {
	lrel := strings.ToLower(rel)
	for _, row := range cats {
		cat := row[0]
		for _, v := range row {
			if strings.ToLower(v) == lrel {
				return cat
			}
		}
	}

	return ""
}
