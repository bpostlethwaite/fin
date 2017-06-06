package finpony

import (
	"log"
	"path"
	"runtime"

	"github.com/BurntSushi/toml"
)

const (
	CONFIG_FILE = "config.toml"
)

type finponyConf struct {
	Project string
	Data    dataInfo
	Creds   credInfo
}

type dataInfo struct {
	Raw     string `toml:"raw_data"`
	SheetId string `toml:"sheet_id"`
}

type credInfo struct {
	GoogleMapsApiKey   string `toml:"google_maps_apikey"`
	GoogleClientSecret string `toml:"google_client_secret"`
}

func GetConfig() (*finponyConf, error) {

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("No caller information")
	}
	configDir := path.Dir(filename)
	configPath := path.Join(configDir, CONFIG_FILE)

	config := &finponyConf{}
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, err
	}

	return config, nil
}

func Config() finponyConf {
	config, err := GetConfig()
	if err != nil {
		log.Fatal(err)
	}
	return *config
}

func ConfigData() dataInfo {
	return Config().Data
}

func ConfigCreds() credInfo {
	return Config().Creds
}
