package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/antzucaro/matchr"
)

const (
	TEXT_URL      = "https://maps.googleapis.com/maps/api/place/textsearch/json"
	COMPARE_CHARS = 15
)

var (
	Lwr = strings.ToLower
)

type Match struct {
	Name      string
	PlaceType string
	Category  string
	Record    Record
	dist      int
}

func (m Match) String() string {
	s := []string{}
	if m.Name != "" {
		s = append(s, fmt.Sprintf("Name: %s", m.Name))
	}

	if m.PlaceType != "" {
		s = append(s, fmt.Sprintf("Place: %s", m.PlaceType))
	}

	if m.Category != "" {
		s = append(s, fmt.Sprintf("Place: %s", m.Category))
	}

	if m.Record != (Record{}) {
		s = append(s, fmt.Sprintf("Record: %s", m.Record))
	}

	return strings.Join(s, " ")
}

func GetCategory(cat string, cats []string) (string, error) {
	lcat := Lwr(cat)
	for _, c := range cats {
		if Lwr(c) == lcat {
			return c, nil
		}
	}

	// We didn't find a category lets recommend one to the user by
	// lexicographic distance
	ms := make([]Match, len(cats))
	for i, c := range cats {
		ms[i] = Match{
			dist: matchr.DamerauLevenshtein(Lwr(c), lcat),
			Name: c,
		}
	}

	sort.Slice(ms, func(i, j int) bool { return ms[i].dist < ms[j].dist })

	// TODO return nothing if distance less than X.
	return "", fmt.Errorf("Category %s not found. Did you mean '%s'",
		cat, ms[0].Name)
}

func SetCategory(q Query) error {
	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	ftxs, err := QueryTxs(q, txs)
	if err != nil {
		return err
	}

	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	if len(tcat) == 0 {
		return fmt.Errorf("Category Sheet is empty.")
	}

	cat, err := GetCategory(q.Val, catsFromTable(tcat))
	if err != nil {
		return err
	}

	for i, _ := range ftxs {
		ftxs[i].Category = cat
	}

	txs = AppendDedupeSort(txs, ftxs)
	return store.WriteTransactionTable(txs)
}

func AddPlace(place, category string) error {
	store := NewStore(ConfigData().SheetId)
	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	for i, cat := range tcat {
		name := cat[0]
		if category == name {
			if len(cat) == 1 {
				tcat[i] = append(tcat[i], place)

			} else {
				for j := 1; j < len(cat); j++ {
					if cat[j] == place {
						// already exists
						return nil
					}
				}
				// doesn't exist
				tcat[i] = append(tcat[i], place)
			}
		}
	}
	return store.WriteCategoryTable(tcat)
}

func AddCat(catName string) error {
	store := NewStore(ConfigData().SheetId)
	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	// If present skip.
	for _, cat := range tcat {
		if Lwr(catName) == Lwr(cat[0]) {
			return nil
		}
	}

	tcat = append(tcat, []string{catName})
	sort.Slice(tcat, func(i, j int) bool { return tcat[i][0] < tcat[j][0] })

	return store.WriteCategoryTable(tcat)
}

func RmCat(catName string) error {

	if catName == UNCATEGORIZED {
		return fmt.Errorf("Cannot remove category %s. "+
			"It is a required category.", UNCATEGORIZED)
	}

	store := NewStore(ConfigData().SheetId)
	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	newCats := [][]string{}
	for _, cat := range tcat {
		if Lwr(catName) == Lwr(cat[0]) {
			continue
		} else {
			newCats = append(newCats, cat)
		}
	}

	newTxs := []Record{}
	for _, tx := range txs {
		if Lwr(catName) == Lwr(tx.Category) {
			tx.Category = UNCATEGORIZED
		}
		newTxs = append(newTxs, tx)
	}

	err = store.WriteCategoryTable(newCats)
	if err != nil {
		return err
	}

	return store.WriteTransactionTable(newTxs)
}

func MvCat(fromCat, toCat string) error {

	if fromCat == UNCATEGORIZED {
		return fmt.Errorf("Cannot move category %s. "+
			"It is a required category.", UNCATEGORIZED)
	}

	store := NewStore(ConfigData().SheetId)
	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	moved := []string{}
	for _, cat := range tcat {
		if Lwr(fromCat) == Lwr(cat[0]) {
			moved = cat
			moved[0] = toCat
		}
	}

	if len(moved) == 0 {
		return fmt.Errorf("Cannot move none existent category '%s'", fromCat)
	}

	// Remove mention of any from or to. We write the previously move category
	// overwriting any Google Places names.
	newCats := [][]string{}
	for _, cat := range tcat {
		if Lwr(toCat) == Lwr(cat[0]) || Lwr(fromCat) == Lwr(cat[0]) {
			continue
		} else {
			newCats = append(newCats, cat)
		}
	}

	newCats = append(newCats, moved)
	sort.Slice(newCats, func(i, j int) bool { return newCats[i][0] < newCats[j][0] })

	newTxs := []Record{}
	for _, tx := range txs {
		if Lwr(fromCat) == Lwr(tx.Category) {
			tx.Category = toCat
		}
		newTxs = append(newTxs, tx)
	}

	err = store.WriteCategoryTable(newCats)
	if err != nil {
		return err
	}

	return store.WriteTransactionTable(newTxs)
}

func catsFromTable(rows [][]string) []string {
	cats := make([]string, len(rows))
	for i, r := range rows {
		cats[i] = r[0]
	}
	return cats
}

func catMapFromTable(rows [][]string) map[string]bool {
	cats := make(map[string]bool, len(rows))
	for _, r := range rows {
		cats[r[0]] = true
	}
	return cats
}
