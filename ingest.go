package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
)

func readRawTx() ([]Record, error) {
	records := []Record{}
	files, err := ioutil.ReadDir(ConfigData().Raw)
	if err != nil {
		return []Record{}, err
	}

	rowTemplates := []RowIndicies{
		// RBC
		RowIndicies{
			Date:       2,
			Name:       4,
			Dollar:     6,
			DateFormat: "1/2/2006",
		},
	}

	for _, file := range files {
		fpath := path.Join(ConfigData().Raw, file.Name())
		f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return []Record{}, err
		}
		defer f.Close()

		r := csv.NewReader(f)
		r.TrimLeadingSpace = true
		r.FieldsPerRecord = -1

		rows, err := r.ReadAll()
		if err != nil {
			return []Record{}, err
		}

		// TODO header detection. Strip header
		rows = rows[1:]

		var rerr error
	NextTemplate:
		for _, ri := range rowTemplates {
			recs := []Record{}
			for _, row := range rows {
				rec, err := ri.RecordFromRow(row)
				if err != nil {
					rerr = err
					continue NextTemplate
				}
				recs = append(recs, rec)
			}
			records = append(records, recs...)
			break
		}

		if len(records) == 0 {
			return []Record{}, rerr
		}
	}

	return records, nil
}

func toUpper(recs []Record) []Record {
	frecs := []Record{}

	for _, r := range recs {
		r.Name = strings.ToUpper(r.Name)
		frecs = append(frecs, r)
	}

	return frecs
}

func unEscape(recs []Record) []Record {
	frecs := []Record{}
	replacer := strings.NewReplacer("&amp;", "&")
	for _, r := range recs {
		r.Name = replacer.Replace(r.Name)
		frecs = append(frecs, r)
	}

	return frecs
}

func filterDuplicates(recs []Record) []Record {
	seen := make(map[string]bool)
	frecs := []Record{}

	for _, rec := range recs {
		k := rec.Key()
		if _, ok := seen[k]; !ok {
			seen[k] = true
			frecs = append(frecs, rec)
		}
	}

	return frecs
}

func Ingest() error {

	// injest raw input transactions
	rawtxs, err := readRawTx()
	if err != nil {
		return err
	}

	rawtxs = unEscape(rawtxs)
	rawtxs = toUpper(rawtxs)

	store := NewStore()

	// read in transactions held in Google Docs
	spreadsheetId := ConfigData().SheetId
	readRange := TX_TABLE // all values

	txs, err := store.ReadTransactionTable(spreadsheetId, readRange)
	if err != nil {
		return err
	}

	// append sort and dedupe
	txs = append(txs, rawtxs...)
	txs = filterDuplicates(txs)
	sort.Sort(ByDate(txs))

	// write back to sheet
	writeRange := readRange
	return store.WriteTransactionTable(spreadsheetId, writeRange, txs)
}
