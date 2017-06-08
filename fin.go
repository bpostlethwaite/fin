package main

import (
	"fmt"
	"os"
	"path"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	UNCATEGORIZED = "Uncategorized"
)

var (
	DEFAULT_CONFIG = path.Join(os.Getenv("HOME"), ".fin.toml")

	app         = kingpin.New("fin", "Financial reporting from the command-line")
	DEBUG       = app.Flag("debug", "Enable debug mode").Bool()
	CONFIG_FILE = app.Flag("config", "Config file path").Default(DEFAULT_CONFIG).String()

	query       = app.Command("query", "Query transactions")
	queryName   = query.Flag("name", "Name or partial name of transaction").String()
	queryExpr   = query.Flag("expr", "Regex to match against transaction").String()
	queryCat    = query.Flag("cat", "transactions with matching category").String()
	querySearch = query.Flag("search", "recommend categories for transactions").Bool()
	queryPlaces = query.Flag("places", "print unmatched Google Places hits").Bool()

	raw = app.Command("raw", "Pull raw data from bank accounts")

	ingest = app.Command("ingest", "Ingest raw tx data into the system")

	apply     = app.Command("apply", "Apply categories to transactions.")
	applyFile = apply.Flag("file", "Apply transactions from file. Useful for applying recommendations.").String()
	applyName = apply.Flag("name", "Name or partial name of transaction.").String()
	applyExpr = apply.Flag("expr", "Regex to match against transaction.").String()
	applyCat  = apply.Arg("category", "Name of category.").String()

	report = app.Command("report", "Generate reports")

	clear      = app.Command("clear", "Clears a sheet. Designed for testing.")
	clearSheet = clear.Arg("sheet", "Name of sheet to clear.").String()
)

func main() {

	var err error
	var out []string

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch cmd {

	case raw.FullCommand():
		err = RawBank()

	case ingest.FullCommand():
		err = Ingest()

	case query.FullCommand():
		if *querySearch {
			out, err = TxStringify(CatSearch())
		} else if *queryPlaces {
			out, err = MatchStringify(UnmatchedPlaceSearch())
		} else {
			out, err = TxStringify(QueryTable(Query{
				Name: *queryName,
				Expr: *queryExpr,
				Cat:  *queryCat,
			}))
		}

	case apply.FullCommand():
		if *applyFile != "" {
			err = IngestFile(*applyFile)
		} else {
			err = Categorize(Query{
				Name: *applyName,
				Expr: *applyExpr,
				Cat:  *applyCat,
			})
		}

	case report.FullCommand():
		err = GenerateReports()

	case clear.FullCommand():
		store := NewStore(ConfigData().SheetId)
		err = store.ClearSheet(*clearSheet)
	}

	if err != nil {
		app.Fatalf(err.Error())
	}

	PrintLines(out)
}

func PrintLines(ss []string) {
	for _, s := range ss {
		fmt.Println(s)
	}
}

func TxStringify(txs []Record, err error) ([]string, error) {
	ss := []string{}
	for _, tx := range txs {
		ss = append(ss, tx.String())
	}
	return ss, err
}

func MatchStringify(ms []Match, err error) ([]string, error) {
	ss := []string{}
	for _, m := range ms {
		ss = append(ss, m.String())
	}
	return ss, err
}
