package app

import (
	"flag"
	"io"
	"log"

	"github.com/spf13/cobra"
	"github.com/stepdc/podacrobat/cmd/app/config"
	"github.com/stepdc/podacrobat/pkg/acrobat"
)

func NewAcrobatCommand(out io.Writer) *cobra.Command {
	app := &config.PodAcrobat{}
	cmd := &cobra.Command{
		Use:   "podacrobat",
		Short: "podacrobat",
		Long:  "podacrobat",
		Run: func(cmd *cobra.Command, args []string) {
			err := Run(app)
			if err != nil {
				log.Printf("%v", err)
			}
		},
	}
	cmd.SetOutput(out)

	flags := cmd.Flags()
	flags.AddGoFlagSet(flag.CommandLine)
	app.AddFlags(flags)

	return cmd
}

func Run(app *config.PodAcrobat) error {
	return acrobat.Run(app)
}
