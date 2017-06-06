package main

import (
	"encoding/csv"
	"io/ioutil"
	"os"
	"path"
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

func Ingest() error {

	// injest raw input transactions
	rawtxs, err := readRawTx()
	if err != nil {
		return err
	}

	rawtxs = unEscape(rawtxs)
	rawtxs = toUpper(rawtxs)

	store := NewStore(ConfigData().SheetId)

	// read in transactions held in Google Docs

	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	// Append sort and dedupe. Always prepend rawtxs before txs so we keep
	// changes in txs not present in the raw data itself when filtering.
	// Filtering takes the more "recent" matching record.
	txs = AppendDedupeSort(rawtxs, txs)

	for i, _ := range txs {
		if txs[i].Category == "" {
			txs[i].Category = UNCATEGORIZED
		}
	}

	// write back to sheet
	return store.WriteTransactionTable(txs)
}
