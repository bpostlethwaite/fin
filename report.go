package main

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	DATAVAR = "window.FinData"
)

type BasePage struct {
	Title string
	Data  Data
}

type Data struct {
	DataVar     template.JS
	Months      []AggCat
	RecentMonth template.JS
}

type AggCat struct {
	month  time.Month
	Month  template.JS
	Labels []string
	Values []float64
	Colors []string
}

type AggPair struct {
	Label string
	Value float64
}

type AggCatRef struct {
	Month  template.JS
	Labels template.JS
	Values template.JS
	Colors template.JS
}

func (ac AggCat) ToPairs() []AggPair {
	ap := []AggPair{}
	for i := 0; i < len(ac.Labels); i++ {
		ap = append(ap, AggPair{ac.Labels[i], ac.Values[i]})
	}

	return ap
}

func (ac AggCat) ToRefs() AggCatRef {
	return AggCatRef{
		Month:  template.JS(ac.Month),
		Labels: template.JS(fmt.Sprintf("%s.%s.labels", DATAVAR, ac.Month)),
		Values: template.JS(fmt.Sprintf("%s.%s.values", DATAVAR, ac.Month)),
		Colors: template.JS(fmt.Sprintf("%s.%s.colors", DATAVAR, ac.Month)),
	}
}

func sortCatsValsByLabel(cats []string, vals []float64) {
	ap := []AggPair{}
	for i := 0; i < len(cats); i++ {
		ap = append(ap, AggPair{cats[i], vals[i]})
	}
	sort.Slice(ap, func(i, j int) bool { return ap[i].Label < ap[j].Label })

	for i, a := range ap {
		cats[i] = a.Label
		vals[i] = a.Value
	}
}

func sortCatsValsByValue(cats []string, vals []float64) {
	ap := []AggPair{}
	for i := 0; i < len(cats); i++ {
		ap = append(ap, AggPair{cats[i], vals[i]})
	}
	sort.Slice(ap, func(i, j int) bool { return ap[i].Value >= ap[j].Value })

	for i, a := range ap {
		cats[i] = a.Label
		vals[i] = a.Value
	}
}

func aggregateCategoryExpenses(txs []Record) ([]string, []float64) {
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

	return cats, vals
}

func groupByMonth(txs []Record) map[time.Month][]Record {
	mmap := map[time.Month][]Record{}
	for _, tx := range txs {
		rxs := mmap[tx.Date.Month()]
		mmap[tx.Date.Month()] = append(rxs, tx)
	}

	return mmap
}

func GenerateReports() error {

	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	txs = filter(txs, func(tx Record, i int) bool { return tx.Dollar <= 0. })
	cmonths := []AggCat{}

	// We want to pair a particular category with the same color month to month.
	// To do so create a category=>color map before chopping data into months.
	cats, vals := aggregateCategoryExpenses(txs)
	sortCatsValsByValue(cats, vals)
	cmap := make(map[string]string, len(cats))
	for i, c := range cats {
		cmap[c] = COLORS[i]
	}

	for k, v := range groupByMonth(txs) {
		cats, vals := aggregateCategoryExpenses(v)
		sortCatsValsByValue(cats, vals)
		colors := values(cmap, cats)
		cmonths = append(cmonths, AggCat{
			k, template.JS(k.String()), cats, vals, colors,
		})
	}

	sort.Slice(cmonths, func(i, j int) bool {
		return cmonths[i].month < cmonths[j].month
	})

	base := BasePage{
		Title: Config().Project,
		Data: Data{
			DataVar:     template.JS(DATAVAR),
			Months:      cmonths,
			RecentMonth: cmonths[len(cmonths)-1].Month,
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

func values(m map[string]string, ar []string) []string {
	vals := []string{}
	for _, k := range ar {
		if v, ok := m[k]; ok {
			vals = append(vals, v)
		}
	}
	return vals
}
