package main

import (
	"html/template"
	"math"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type BasePage struct {
	Title string
	Data  Data
}

type Data struct {
	DataVar     template.JS
	Months      []AggCat
	All         AggCat
	RecentMonth template.JS
}

type AggCat struct {
	month  time.Month
	Month  template.JS
	Txs    []Record
	CatTxs []CatTx
	Labels []string
	Values []float64
	Colors []string
}

type AggPair struct {
	Label string
	Value float64
	Color string
}

func (ac AggCat) ToPairs() []AggPair {
	ap := []AggPair{}
	for i := 0; i < len(ac.Labels); i++ {
		c := ""
		if len(ac.Colors) > i {
			c = ac.Colors[i]
		}
		ap = append(ap, AggPair{ac.Labels[i], ac.Values[i], c})
	}

	return ap
}

type CatTx struct {
	Category template.JS
	Color    template.JS
	Txs      []Record
}

func txsByCat(txs []Record, cmap map[string]string) []CatTx {
	cts := []CatTx{}
	iseen := map[string]int{}

	for _, tx := range txs {
		i, ok := iseen[tx.Category]
		if !ok {
			c, _ := cmap[tx.Category]
			cts = append(cts, CatTx{
				Category: template.JS(tx.Category),
				Color:    template.JS(c),
				Txs:      []Record{tx},
			})
			iseen[tx.Category] = len(cts) - 1
		} else {
			cts[i].Txs = append(cts[i].Txs, tx)
		}
	}

	return cts
}

func sortCatsValsByLabel(cats []string, vals []float64) {
	ap := []AggPair{}
	for i := 0; i < len(cats); i++ {
		ap = append(ap, AggPair{Label: cats[i], Value: vals[i]})
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
		ap = append(ap, AggPair{Label: cats[i], Value: vals[i]})
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

	funcMap := template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"abs": math.Abs,
	}

	store := NewStore(ConfigData().SheetId)
	txs, err := store.ReadTransactionTable()
	if err != nil {
		return err
	}

	txs = filter(txs, func(tx Record, i int) bool { return tx.Dollar <= 0. })

	// We want to pair a particular category with the same color month to month.
	// To do so create a category=>color map before chopping data into months.
	cats, vals := aggregateCategoryExpenses(txs)
	sortCatsValsByValue(cats, vals)
	cmap := make(map[string]string, len(cats))
	for i, c := range cats {
		cmap[c] = COLORS[i]
	}
	colors := values(cmap, cats)
	all := AggCat{
		Labels: cats,
		Values: vals,
		Colors: colors,
		CatTxs: txsByCat(txs, cmap),
		Txs:    txs,
	}

	// Monthly Data
	cmonths := []AggCat{}
	for k, v := range groupByMonth(txs) {
		sort.Sort(ByDate(v))
		cats, vals := aggregateCategoryExpenses(v)
		sortCatsValsByValue(cats, vals)
		colors := values(cmap, cats)
		cmonths = append(cmonths, AggCat{
			month:  k,
			Month:  template.JS(k.String()),
			Labels: cats,
			Values: vals,
			Colors: colors,
			CatTxs: txsByCat(v, cmap),
			Txs:    v,
		})
	}
	sort.Slice(cmonths, func(i, j int) bool {
		return cmonths[i].month < cmonths[j].month
	})

	base := BasePage{
		Title: Config().Project,
		Data: Data{
			All:         all,
			Months:      cmonths,
			RecentMonth: cmonths[len(cmonths)-1].Month,
		},
	}

	pattern := filepath.Join(Config().ProjectPath, "templates", "*.tmpl")
	templates := template.Must(
		template.New("cheese").Funcs(funcMap).ParseGlob(pattern),
	)

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
