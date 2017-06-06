package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

const (
	GOOGLE_SPREADSHEET_URL = "https://www.googleapis.com/auth/spreadsheets"
	TX_TABLE               = "Transactions"
	CAT_TABLE              = "Categories"
	TX_CAT_TABLE           = "TransactionCategories"
)

type Store struct {
	Service *sheets.Service
}

func (s *Store) ReadTransactionTable(sheetId, readRange string) ([]Record, error) {

	srv := s.Service
	resp, err := srv.Spreadsheets.Values.Get(sheetId, readRange).Do()
	if err != nil {
		return []Record{}, fmt.Errorf("Unable to retrieve data from sheet. %v", err)
	}

	ri := RowIndicies{0, 1, 2, DATE_FORMAT}

	txs := []Record{}
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			rec, err := ri.RecordFromRow(stringsFromInteraces(row))
			if err != nil {
				return []Record{}, err
			}
			txs = append(txs, rec)
		}
	} else {
		return []Record{}, nil
	}

	return txs, nil
}

func (s *Store) WriteTransactionTable(sheetId, writeRange string, txs []Record) error {

	err := s.ClearTransactionTable(sheetId, writeRange)
	if err != nil {
		return err
	}

	values := sheets.ValueRange{
		Values: To(txs).Rows(),
	}

	_, err = s.Service.Spreadsheets.Values.Update(
		sheetId, writeRange, &values,
	).ValueInputOption("USER_ENTERED").Do()

	return err
}

func (s *Store) ClearTransactionTable(sheetId, clearRange string) error {
	opts := sheets.ClearValuesRequest{}
	_, err := s.Service.Spreadsheets.Values.Clear(sheetId, clearRange, &opts).Do()
	return err
}

func NewStore() *Store {
	ctx := context.Background()
	b, err := ioutil.ReadFile(ConfigCreds().GoogleClientSecret)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, GOOGLE_SPREADSHEET_URL)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}

	return &Store{srv}
}

func stringsFromInteraces(si []interface{}) []string {
	ss := []string{}
	for _, s := range si {
		ss = append(ss, s.(string))
	}
	return ss
}
