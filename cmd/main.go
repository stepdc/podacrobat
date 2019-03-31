package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/stepdc/podacrobat/cmd/app"
)

func main() {
	out := os.Stdout
	cmd := app.NewAcrobatCommand(out)
	flag.CommandLine.Parse([]string{})
	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
