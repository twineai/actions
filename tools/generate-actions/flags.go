package main

import (
	"github.com/namsral/flag"
)

var (
	FlagActionDir string
)

func init() {
	flag.StringVar(
		&FlagActionDir, "action-dir", "",
		"The path where actions should be created")
}
