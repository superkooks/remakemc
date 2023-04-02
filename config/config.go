package config

import (
	"errors"
	"os"

	_ "embed"

	"gopkg.in/yaml.v3"
)

// Documentation for AppConfig resides in the default config file
type AppConfig struct {
	PublicServer bool
	ServePprof   bool

	Client struct {
		DefaultHeight int
		DefaultWidth  int
	}
}

var App AppConfig

//go:embed defaultConfig.yaml
var defaultConfig []byte

func ParseConfig() {
	f, err := os.Open("config.yaml")
	if errors.Is(err, os.ErrNotExist) {
		CreateDefaultConfig()
		return
	} else if err != nil {
		panic(err)
	}
	defer f.Close()

	// We can use yaml to parse json with comments, as yaml is a strict superset
	// of json with comments
	d := yaml.NewDecoder(f)
	err = d.Decode(&App)
	if err != nil {
		panic(err)
	}
}

func CreateDefaultConfig() {
	f, err := os.Create("config.yaml")
	if err != nil {
		panic(err)
	}

	f.Write(defaultConfig)
	f.Close()

	ParseConfig()
}
