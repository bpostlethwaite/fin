package finpony

import (
	"fmt"
	"strconv"
	"time"
)

const DATE_FORMAT = "2006-1-2"

type RowIndicies struct {
	Date       int64
	Name       int64
	Dollar     int64
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

	return Record{
		Date:   date,
		Name:   name,
		Dollar: dollar,
	}, nil
}

type Record struct {
	Date   time.Time `csv:"Date"`
	Name   string    `csv:"Description"`
	Dollar float64   `csv:"Value"`
	Type   string
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
		r.Dollar, // strconv.FormatFloat(r.Dollar, 'f', -1, 64)
	}
}

type ByDate []Record

func (a ByDate) Len() int           { return len(a) }
func (a ByDate) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByDate) Less(i, j int) bool { return a[i].Date.Before(a[j].Date) }

type To []Record

func (rs To) Rows() [][]interface{} {
	rows := make([][]interface{}, len(rs))
	for i, r := range rs {
		rows[i] = r.Row()
	}

	return rows
}
