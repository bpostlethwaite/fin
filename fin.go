package main

import (
	"fmt"
	"log"
	"os"
	"path"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	DEFAULT_CONFIG = path.Join(os.Getenv("HOME"), ".fin.toml")

	app         = kingpin.New("fin", "Financial reporting from the command-line.")
	debug       = app.Flag("debug", "Enable debug mode.").Bool()
	CONFIG_FILE = app.Flag("config", "Config file path").Default(DEFAULT_CONFIG).String()

	raw = app.Command("raw", "Pull raw data from bank accounts")

	ingest = app.Command("ingest", "ingest raw tx data into the system")

	cat = app.Command("cat", "Categorize transactions")
)

func main() {

	var err error
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))
	switch cmd {

	case raw.FullCommand():
		err = fmt.Errorf("%s not implemented", cmd)

	case ingest.FullCommand():
		err = Ingest()

	case cat.FullCommand():
		err = fmt.Errorf("%s not implemented", cmd)
	}

	if err != nil {
		log.Fatal(err)
	}
}
