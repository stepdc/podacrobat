package app

import (
	"flag"
	"io"

	"github.com/stepdc/podacrobat/pkg/acrobat"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/stepdc/podacrobat/cmd/app/config"

	"k8s.io/kubernetes/pkg/kubectl/util/logs"
)

func NewAcrobatCommand(out io.Writer) *cobra.Command {
	app := &config.PodAcrobat{}
	cmd := &cobra.Command{
		Use:   "podacrobat",
		Short: "podacrobat",
		Long:  "podacrobat",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InitLogs()
			defer logs.FlushLogs()
			err := func(app *config.PodAcrobat) error {
				return acrobat.Run(app)
			}(app)
			if err != nil {
				glog.Errorf("%v", err)
			}
		},
	}
	cmd.SetOutput(out)

	flags := cmd.Flags()
	flags.AddGoFlagSet(flag.CommandLine)
	app.AddFlags(flags)

	return cmd
}
