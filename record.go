package main

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

const DATE_FORMAT = "2006-1-2"

type RowIndicies struct {
	Date       int
	Name       int
	Dollar     int
	Category   int
	DateFormat string
}

func (ri RowIndicies) RecordFromRow(row []string) (Record, error) {

	date, err := time.Parse(ri.DateFormat, row[ri.Date])
	if err != nil {
		return Record{}, err
	}

	dollar, err := strconv.ParseFloat(row[ri.Dollar], 64)
	if err != nil {
		return Record{}, err
	}

	name := row[ri.Name]

	var category string
	if ri.Category != 0 && len(row) > ri.Category {
		category = row[ri.Category]
	}

	return Record{
		Date:     date,
		Name:     name,
		Dollar:   dollar,
		Category: category,
	}, nil
}

type Record struct {
	Date     time.Time
	Name     string
	Dollar   float64
	Category string
}

func (r Record) String() string {
	s := fmt.Sprintf("%s \"%s\" %s",
		r.Date.Format(DATE_FORMAT),
		r.Name,
		strconv.FormatFloat(r.Dollar, 'f', -1, 64))

	if r.Category == "" {
		return s
	}

	return fmt.Sprintf("%s %s", s, r.Category)
}

func (r Record) Key() string {
	date := r.Date.Format(DATE_FORMAT)
	name := r.Name
	amount := strconv.FormatFloat(r.Dollar, 'f', -1, 64)
	return fmt.Sprintf("%s+%s+%s", date, name, amount)
}

func (r Record) Row() []interface{} {
	return []interface{}{
		r.Date.Format(DATE_FORMAT),
		r.Name,
		r.Dollar,
		r.Category,
	}
}

type ByDate []Record

func (a ByDate) Len() int {
	return len(a)
}

func (a ByDate) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByDate) Less(i, j int) bool {
	if a[i].Date.Equal(a[j].Date) {
		if a[i].Name == a[j].Name {
			return a[i].Dollar >= a[j].Dollar // dollars are negative
		}
		return a[i].Name < a[j].Name
	}
	return a[i].Date.Before(a[j].Date)
}

type To []Record

func (rs To) Rows() [][]interface{} {
	rows := make([][]interface{}, len(rs))
	for i, r := range rs {
		rows[i] = r.Row()
	}

	return rows
}

func AppendDedupeSort(a, b []Record) []Record {
	c := filterDuplicates(append(a, b...))
	sort.Sort(ByDate(c))
	return c
}

func filterDuplicates(recs []Record) []Record {
	seen := make(map[string]bool)
	frecs := []Record{}

	for i := len(recs) - 1; i >= 0; i-- {
		rec := recs[i]
		k := rec.Key()
		if _, ok := seen[k]; !ok {
			seen[k] = true
			frecs = append(frecs, rec)
		}
	}

	return frecs
}
