package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
)

const (
	CONFIG_ENV_VAR = "FIN_CONFIG_PATH"
)

func readRawTx() ([]Record, error) {
	records := []Record{}
	files, err := ioutil.ReadDir(ConfigData().Raw)
	if err != nil {
		return nil, err
	}

	rowTemplates := []RowIndicies{
		// RBC
		RowIndicies{
			Date:       2,
			Name:       4,
			Dollar:     6,
			DateFormat: "1/2/2006",
		},

		// App format
		DefaultRowIndicies(),
	}

	for _, file := range files {
		fpath := path.Join(ConfigData().Raw, file.Name())
		txs, err := ReadFromFile(fpath, rowTemplates)
		if err != nil {
			return nil, err
		}

		records = append(records, txs...)
	}

	return records, nil
}

func ReadFromFile(fpath string, tmpls []RowIndicies) ([]Record, error) {
	f, err := os.OpenFile(fpath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true
	r.FieldsPerRecord = -1

	rows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	txs := []Record{}
	rlen := len(rows)
	var rerr error
NextTemplate:
	for _, ri := range tmpls {
		recs := []Record{}
		for i, row := range rows {
			rec, err := ri.RecordFromRow(row)
			skipErr := rlen > 1 && i == 0
			if err != nil {
				if skipErr {
					continue
				}
				rerr = err
				continue NextTemplate
			}

			recs = append(recs, rec)
		}
		txs = append(txs, recs...)
		break
	}

	if len(txs) == 0 {
		return nil, rerr
	}

	return txs, nil
}

func IngestFile(fpath string) error {
	newtxs, err := ReadFromFile(fpath, []RowIndicies{DefaultRowIndicies()})
	if err != nil {
		return err
	}

	newtxs = unEscape(newtxs)
	newtxs = toUpper(newtxs)

	store := NewStore(ConfigData().SheetId)

	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	// newtxs come after as we are overwriting matching txs
	txs = AppendDedupeSort(txs, newtxs)

	for i, _ := range txs {
		if txs[i].Category == "" {
			txs[i].Category = UNCATEGORIZED
		}
	}

	// write back to sheet
	return store.WriteTransactionTable(txs)
}

func IngestCache() error {

	// injest raw input transactions
	rawtxs, err := readRawTx()
	if err != nil {
		return err
	}

	rawtxs = unEscape(rawtxs)
	rawtxs = toUpper(rawtxs)

	store := NewStore(ConfigData().SheetId)

	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	tcat, err := store.ReadCategoryTable()
	if err != nil {
		return err
	}

	// Append sort and dedupe. Always prepend rawtxs before txs so we keep
	// changes in txs not present in the raw data itself when filtering.
	// Filtering takes the more "recent" matching record.
	txs = AppendDedupeSort(rawtxs, txs)
	cats := catMapFromTable(tcat)
	for i, _ := range txs {
		if _, ok := cats[txs[i].Category]; !ok {
			txs[i].Category = UNCATEGORIZED
		}
	}

	// write back to sheet
	return store.WriteTransactionTable(txs)
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

func RunScripts(script string) error {
	rawDir := path.Join(Config().ProjectPath, "rawbank")
	nightwatch := path.Join(rawDir, "node_modules/.bin/nightwatch")
	scriptarg := ""
	if script != "" {
		scriptarg = fmt.Sprintf("--test %s", script)
	}

	configEnv := fmt.Sprintf("%s=%s", CONFIG_ENV_VAR, *CONFIG_FILE)

	ctx := context.Background()
	cmd := exec.CommandContext(ctx, nightwatch, scriptarg)

	cmd.Dir = rawDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), configEnv)

	return cmd.Run()
}
