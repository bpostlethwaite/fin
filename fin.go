package main

import (
	"fmt"
	"log"
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

	query     = app.Command("query", "Query transactions")
	queryName = query.Flag("name", "Name or partial name of transaction").String()
	queryExpr = query.Flag("expr", "Regex to match against transaction").String()
	queryCat  = query.Flag("cat", "transactions with matching category").String()

	raw = app.Command("raw", "Pull raw data from bank accounts")

	ingest = app.Command("ingest", "Ingest raw tx data into the system")

	assign     = app.Command("assign", "Assign categories to transactions ")
	assignName = assign.Flag("name", "Name or partial name of transaction").String()
	assignExpr = assign.Flag("expr", "Regex to match against transaction").String()
	assignCat  = assign.Arg("category", "Name of category").Required().String()
)

func main() {

	var err error
	var txs []Record

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch cmd {

	case raw.FullCommand():
		err = fmt.Errorf("%s not implemented", cmd)

	case ingest.FullCommand():
		err = Ingest()

	case query.FullCommand():
		txs, err = QueryTable(Query{
			Name: *queryName,
			Expr: *queryExpr,
			Cat:  *queryCat,
		})

	case assign.FullCommand():
		err = Categorize(Query{
			Name: *assignName,
			Expr: *assignExpr,
			Cat:  *assignCat,
		})
	}

	if err != nil {
		log.Fatal(err)
	}
	if len(txs) > 0 {
		PrintTxs(txs)
	}
}

func PrintTxs(txs []Record) {
	for _, tx := range txs {
		fmt.Println(tx)
	}
}
