package app

import (
	"flag"
	"io"

	"github.com/stepdc/podacrobat/pkg/acrobat"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/kubectl/util/logs"
)

type PodAcrobat struct {
	Config
	Client clientset.Interface
}

func NewAcrobatCommand(out io.Writer) *cobra.Command {
	app := &PodAcrobat{}
	cmd := &cobra.Command{
		Use:   "podacrobat",
		Short: "podacrobat",
		Long:  "podacrobat",
		Run: func(cmd *cobra.Command, args []string) {
			logs.InitLogs()
			defer logs.FlushLogs()
			err := func(app *PodAcrobat) error {
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

func (pa *PodAcrobat) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&pa.Policy, "policy", string(PodsCount), "nodes filter policy(use \"podscount\" for test)")
	fs.StringVar(&pa.LowerThreshold, "lowerthreshold", "30", "lower threshold")
	fs.StringVar(&pa.UpperThreshold, "upperthreshold", "50", "upper threshold")
}
