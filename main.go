package main

import (
	"github.com/machmum/depp/config"
	"github.com/machmum/depp/api"
)

func main() {

	path := "./config.local.yaml"

	// get config
	cfg, err := config.Load(path)
	if err != nil {
		panic(err.Error())
	}

	// run engine
	api.Start(cfg)
}
