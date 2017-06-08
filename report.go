package main

import (
	"html/template"
	"math"
	"os"
	"path/filepath"
)

type BasePage struct {
	Title string
	Data  Data
	Pie   Pie
}

type Pie struct {
	Values template.JS
	Labels template.JS
	Title  string
}

type Data struct {
	Ref        template.JS
	Categories StringData
	Dollars    NumberData
}

type NumberData struct {
	Ref template.JS
	Val []float64
}

type StringData struct {
	Ref template.JS
	Val []string
}

func filterPayments(txs []Record) []Record {
	out := []Record{}
	for _, tx := range txs {
		if tx.Dollar <= 0. {
			out = append(out, tx)
		}
	}

	return out
}

func GenerateReports() error {

	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	txs = filterPayments(txs)

	catmap := map[string]float64{}
	for _, tx := range txs {
		catmap[tx.Category] = catmap[tx.Category] + math.Abs(tx.Dollar)
	}

	cats := []string{}
	vals := []float64{}
	for k, v := range catmap {
		cats = append(cats, k)
		vals = append(vals, v)
	}

	data := Data{
		Ref: template.JS("window.FinData"),
		Categories: StringData{
			Ref: template.JS("window.FinData.cat"),
			Val: cats,
		},
		Dollars: NumberData{
			Ref: template.JS("window.FinData.amount"),
			Val: vals,
		},
	}

	base := BasePage{
		Title: "fin",
		Data:  data,
		Pie: Pie{
			Title:  "finpie",
			Labels: data.Categories.Ref,
			Values: data.Dollars.Ref,
		},
	}

	pattern := filepath.Join(Config().ProjectPath, "templates", "*.tmpl")
	templates := template.Must(template.ParseGlob(pattern))

	reportPath := filepath.Join(ConfigData().Reports, "report.html")
	f, err := os.OpenFile(reportPath, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return templates.ExecuteTemplate(f, "base", base)
}
