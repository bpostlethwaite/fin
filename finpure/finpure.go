package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/bpostlethwaite/finpony"
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
}

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

func toUpper(recs []finpony.Record) []finpony.Record {
	frecs := []finpony.Record{}

	for _, r := range recs {
		r.Name = strings.ToUpper(r.Name)
		frecs = append(frecs, r)
	}

	return frecs
}

func unEscape(recs []finpony.Record) []finpony.Record {
	frecs := []finpony.Record{}
	replacer := strings.NewReplacer("&amp;", "&")
	for _, r := range recs {
		r.Name = replacer.Replace(r.Name)
		frecs = append(frecs, r)
	}

	return frecs
}

func filterDuplicates(recs []finpony.Record) []finpony.Record {
	seen := make(map[string]bool)
	frecs := []finpony.Record{}

	for _, rec := range recs {
		k := rec.Key()
		if _, ok := seen[k]; !ok {
			seen[k] = true
			frecs = append(frecs, rec)
		}
	}

	return frecs
}

func extractName(n string) *Match {
	matches := withAddr.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{matches[1], matches[2]}
	}

	matches = withDigits.FindStringSubmatch(n)
	if len(matches) > 2 {
		return &Match{matches[1], matches[2]}
	}

	for _, c := range cities {
		withCity := regexp.MustCompile(fmt.Sprintf("(.*)\\s+(%s.*)", c))
		matches := withCity.FindStringSubmatch(n)
		if len(matches) > 2 {
			return &Match{matches[1], matches[2]}
		}
	}

	return nil
}

func printTxs(txs []finpony.Record) {
	for _, tx := range txs {
		fmt.Println(tx)
	}
}

func main() {

	// injest raw input transactions
	rawtxs, err := readRawTx()
	if err != nil {
		log.Fatal(err)
	}

	rawtxs = unEscape(rawtxs)
	rawtxs = toUpper(rawtxs)

	store := finpony.NewStore()

	// read in transactions held in Google Docs
	spreadsheetId := finpony.ConfigData().SheetId
	readRange := finpony.TX_TABLE // all values

	txs, err := store.ReadTransactionTable(spreadsheetId, readRange)
	if err != nil {
		log.Fatal(err)
	}

	// append sort and dedupe
	txs = append(txs, rawtxs...)
	txs = filterDuplicates(txs)
	sort.Sort(finpony.ByDate(txs))

	// write back to sheet
	writeRange := readRange
	err = store.WriteTransactionTable(spreadsheetId, writeRange, txs)
	if err != nil {
		log.Fatal(err)
	}
}
