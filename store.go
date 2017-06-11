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
)

type Store struct {
	sheetId string
	service *sheets.Service
}

func (s *Store) ReadSheet(readRange string) ([][]string, error) {
	srv := s.service
	resp, err := srv.Spreadsheets.Values.Get(s.sheetId, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve data from sheet. %v", err)
	}

	vals := make([][]string, len(resp.Values))
	for i, row := range resp.Values {
		vals[i] = stringsFromInteraces(row)
	}

	return vals, nil
}

func (s *Store) WriteSheet(writeRange string, vals [][]interface{}) error {
	err := s.ClearSheet(writeRange)
	if err != nil {
		return err
	}

	values := sheets.ValueRange{
		Values: vals,
	}

	_, err = s.service.Spreadsheets.Values.Update(
		s.sheetId, writeRange, &values,
	).ValueInputOption("USER_ENTERED").Do()

	return err
}

func (s *Store) ReadTransactionTable() ([]Record, error) {
	vals, err := s.ReadSheet(TX_TABLE)
	if err != nil {
		return nil, err
	}

	ri := DefaultRowIndicies()

	txs := []Record{}
	for _, row := range vals {
		rec, err := ri.RecordFromRow(row)
		if err != nil {
			return nil, err
		}
		txs = append(txs, rec)
	}

	return txs, nil
}

func (s *Store) WriteTransactionTable(txs []Record) error {
	return s.WriteSheet(TX_TABLE, To(txs).RowsInterface())
}

func (s *Store) ClearSheet(clearRange string) error {
	opts := sheets.ClearValuesRequest{}
	_, err := s.service.Spreadsheets.Values.Clear(s.sheetId, clearRange, &opts).Do()
	return err
}

func (s *Store) ReadCategoryTable() ([][]string, error) {
	vals, err := s.ReadSheet(CAT_TABLE)
	if err != nil {
		return nil, err
	}

	cats := make([][]string, len(vals))
	for i, _ := range vals {
		cats[i] = vals[i]
	}

	return cats, nil
}

func (s *Store) WriteCategoryTable(cats [][]string) error {
	vals := make([][]interface{}, len(cats))
	for i, _ := range cats {
		vals[i] = stringsToInterfaces(cats[i])
	}
	return s.WriteSheet(CAT_TABLE, vals)
}

func NewStore(sheetId string) *Store {
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

	return &Store{
		service: srv,
		sheetId: sheetId,
	}
}

func stringsFromInteraces(si []interface{}) []string {
	ss := []string{}
	for _, s := range si {
		ss = append(ss, s.(string))
	}
	return ss
}

func stringsToInterfaces(si []string) []interface{} {
	ss := []interface{}{}
	for _, s := range si {
		ss = append(ss, s)
	}
	return ss
}
