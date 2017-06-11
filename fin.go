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

type List struct {
	*kingpin.CmdClause
	name  string
	regex string
	cat   string
}

var (
	DEFAULT_CONFIG = path.Join(os.Getenv("HOME"), ".fin.toml")

	// TODO check config outside of kingpin so we can set name as project
	app         = kingpin.New("fin", "Financial reporting from the command-line.")
	CONFIG_FILE = app.Flag("config", "Config file path").Default(DEFAULT_CONFIG).String()

	tx = app.Command("tx", "List and modify transactions")

	list      = List{CmdClause: tx.Command("list", "List transactions.")}
	listname  = list.Flag("name", "List transaction matching name.").String()
	listregex = list.Flag("regex", "List transactions matching regex.").String()
	listcat   = list.Flag("category", "List transactions labeled with category.").String()

	set      = tx.Command("set-category", "Set the category of a transaction.")
	setval   = set.Arg("category", "The category to set.").Required().String()
	setname  = set.Flag("name", "Set category of transactions matching name.").String()
	setregex = set.Flag("regex", "Set category of transactions matching regex.").String()
	setcat   = set.Flag("category", "Set category of transactions labeled with category. Note this effectively 'swaps' categories.").String()

	recommend    = tx.Command("recommend", "Generate an updated list of uncategorized transactions with newly recommended categories.")
	placematches = recommend.Flag("place-misses", "Include list of google place type found in the search that do not match registered categories.").Bool()

	cat = app.Command("category", "List and modify categories.")

	addplace    = cat.Command("add-place", "Add a google place type to category")
	addplaceval = addplace.Arg("place", "The name of place to add to category.").Required().String()
	addplacecat = addplace.Arg("category", "Category in which to add place.").Required().String()

	catadd    = cat.Command("new", "Add a new category into registered categories")
	cataddval = catadd.Arg("category", "Name of category to add into registered categories.").Required().String()

	catrm    = cat.Command("rm", "Remove a category from registered categories and transactions.")
	catrmval = catrm.Arg("category", "Name of category to remove from registered cateogies and transactions.").Required().String()

	catmv     = cat.Command("mv", "Rename a registered category.")
	catmvfrom = catmv.Arg("from", "Name of category to rename").Required().String()
	catmvto   = catmv.Arg("to", "Name of new category").Required().String()

	ingest = app.Command("ingest", "Ingest raw transaction data into the system.")

	ingestfile    = ingest.Command("file", "Ingest a file containing transactions into system.")
	ingestfilevar = ingestfile.Arg("filename", "Path specifying file to ingest").Required().String()

	ingestweb      = ingest.Command("web", "Ingest data from a webite using nightwatch script.")
	ingestscript   = ingestweb.Arg("script", "Name of script to run in nightwatch.").String()
	ingestnoscript = ingestweb.Flag("no-script", "Do not run script. Ingest from cached directory only.").Bool()
	cacheonly      = ingestweb.Flag("cache-only", "Run the script but download transactions to cache directory only. Do not ingest transactions into the system").Bool()

	report = app.Command("report", "Generate reports")

	clear      = app.Command("clear", "Clears a sheet. Designed for testing.")
	clearSheet = clear.Arg("sheet", "Name of sheet to clear.").String()
)

func main() {

	var err error
	var out []string

	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch cmd {

	case list.FullCommand():
		q := Query{
			Name: *listname,
			Expr: *listregex,
			Cat:  *listcat,
		}

		if (q == Query{}) {
			err = argError("list command requires one flag to be set")
		} else {
			out, err = TxStringify(QueryTable(q))
		}

	case set.FullCommand():
		q := Query{
			Name: *setname,
			Expr: *setregex,
			Cat:  *setcat,
			Val:  *setval,
		}

		if (q == Query{}) {
			err = argError("list command requires one flag to be set")
		} else {
			err = SetCategory(q)
		}

	case recommend.FullCommand():
		if *placematches {
			out, err = MatchStringify(PlaceSearch())
		} else {
			out, err = TxStringify(Recommend())
		}

	case addplace.FullCommand():
		err = AddPlace(*addplaceval, *addplacecat)

	case catadd.FullCommand():
		err = AddCat(*cataddval)

	case catrm.FullCommand():
		err = RmCat(*catrmval)

	case catmv.FullCommand():
		err = MvCat(*catmvfrom, *catmvto)

	case ingestfile.FullCommand():
		err = IngestFile(*ingestfilevar)

	case ingestweb.FullCommand():
		if !*ingestnoscript {
			err = RunScripts(*ingestscript)
		}
		if err == nil && !*cacheonly {
			err = IngestCache()
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

func argError(s string) error {
	return fmt.Errorf("%s. '%s help' for usage", Config().Project)
}
