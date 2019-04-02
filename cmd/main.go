package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/stepdc/podacrobat/cmd/app"
)

func main() {
	fmt.Println("fmt start")
	out := os.Stdout
	cmd := app.NewAcrobatCommand(out)
	flag.CommandLine.Parse([]string{})
	if err := cmd.Execute(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
