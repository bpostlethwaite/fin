package main

import (
	"encoding/csv"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"

	"github.com/bpostlethwaite/finpony"
)

var fail = finpony.Fail

func readRawTx() ([]finpony.Record, error) {
	records := []finpony.Record{}
	files, err := ioutil.ReadDir(finpony.ConfigData().Raw)
	if err != nil {
		return []finpony.Record{}, err
	}

	rowTemplates := []finpony.RowIndicies{
		// RBC
		finpony.RowIndicies{
			Date:       2,
			Name:       4,
			Dollar:     6,
			DateFormat: "1/2/2006",
		},
	}

	for _, file := range files {
		fpath := path.Join(finpony.ConfigData().Raw, file.Name())
		f, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			return []finpony.Record{}, err
		}
		defer f.Close()

		r := csv.NewReader(f)
		r.TrimLeadingSpace = true
		r.FieldsPerRecord = -1

		rows, err := r.ReadAll()
		if err != nil {
			return []finpony.Record{}, err
		}

		// TODO header detection. Strip header
		rows = rows[1:]

		var rerr error
	NextTemplate:
		for _, ri := range rowTemplates {
			recs := []finpony.Record{}
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
			return []finpony.Record{}, rerr
		}
	}

	return records, nil
}

func filterDuplicates(recs []finpony.Record) []finpony.Record {
	seen := make(map[string]bool)
	frecs := []finpony.Record{}

	for _, r := range recs {
		k := r.Key()
		if _, ok := seen[k]; !ok {
			seen[k] = true
			frecs = append(frecs, r)
		}
	}

	return frecs
}

func main() {

	// injest raw input transactions
	rawtxs, err := readRawTx()
	if err != nil {
		log.Fatal(err)
	}

	// do basic clean up
	sort.Sort(finpony.ByDate(rawtxs))
	rawtxs = filterDuplicates(rawtxs)

	// separate company names out of descriptions

	store := finpony.NewStore()

	// read in transactions held in Google Docs
	spreadsheetId := finpony.ConfigData().TableId
	readRange := finpony.TX_TABLE // all values

	txs, err := store.ReadTransactionTable(spreadsheetId, readRange)
	if err != nil {
		log.Fatal(err)
	}

	// append sort and dedupe
	txs = append(txs, rawtxs...)
	sort.Sort(finpony.ByDate(txs))
	txs = filterDuplicates(txs)

	// write back to sheet
	writeRange := readRange
	err = store.WriteTransactionTable(spreadsheetId, writeRange, txs)
	if err != nil {
		log.Fatal(err)
	}
}
