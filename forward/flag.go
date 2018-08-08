package gateway

import (
	"flag"
)

var ConfigFile string

func init() {
	flag.StringVar(&ConfigFile, "c", `G:\gotest\src\test2\config.json`, "config file path")
	flag.Parse()
}
