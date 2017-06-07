package main

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"googlemaps.github.io/maps"

	"github.com/antzucaro/matchr"
)

const (
	TEXT_URL      = "https://maps.googleapis.com/maps/api/place/textsearch/json"
	COMPARE_CHARS = 15
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

	cat, err := GetCategory(q.Cat, catsFromTable(cats))
	if err != nil {
		return err
	}

	for i, _ := range ftxs {
		ftxs[i].Category = cat
	}

	txs = AppendDedupeSort(txs, ftxs)
	return store.WriteTransactionTable(txs)
}

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
	ftxs := InternalSearch(txs)

	// Now check Google places to see if there any other recommendations.
	// Pass in all txs updated with Internal search.
	nftxs, err := GooglePlaces(AppendDedupeSort(txs, ftxs), cats)
	if err != nil {
		return nil, err
	}

	// return only updated transactions
	return append(ftxs, nftxs...), nil
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

func catsFromTable(rows [][]string) []string {
	cats := make([]string, len(rows))
	for i, r := range rows {
		cats[i] = r[0]
	}
	return cats
}

func queryUrl(query string) string {
	return fmt.Sprintf("%s?query=%s&key=%s",
		TEXT_URL, query, ConfigCreds().GoogleMapsApiKey)
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

var replDigits *regexp.Regexp = regexp.MustCompile("(.*\\s+)([0-9]+(-?[0-9]+)+)(\\s+.*)")

func GooglePlaces(txs []Record, cats [][]string) ([]Record, error) {

	recs := []Record{}
	// remove weird digits and whatnot
	for i, _ := range txs {
		if txs[i].Category != UNCATEGORIZED {
			continue
		}
		subName := replDigits.ReplaceAllString(txs[i].Name, "$1$4")
		gcat, err := PlaceType(subName)
		if err != nil {
			if strings.Contains(err.Error(), "ZERO_RESULTS") {
				continue
			}
			return nil, err
		}

		cat := getCategoryFromRelated(gcat, cats)
		if cat != "" {
			txs[i].Category = cat
			recs = append(recs, txs[i])
		}
	}

	return recs, nil
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
