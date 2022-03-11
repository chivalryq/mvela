package pkg

import (
	"fmt"
	"os"

	l "github.com/rancher/k3d/v5/pkg/logger"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const VelaCoreVersion = "1.2.2"

type rootFlag struct {
	Debug      bool
	ConfigFile string
}

var (
	flag      rootFlag
	cmdConfig Config
	err       error
	debugMode bool
)

func NewCmdMVela() *cobra.Command {
	rootCmd := cobra.Command{
		Use:   "mvela",
		Short: "mvela is a tool helps run KubeVela in Docker",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			cmdConfig, err = ReadConfig(flag.ConfigFile)

			if err != nil {
				klog.ErrorS(err, "fail to read config file")
				os.Exit(1)
			}
			l.Log().SetLevel(logrus.FatalLevel)
			if flag.Debug {
				l.Log().SetLevel(logrus.DebugLevel)
				debugMode = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("reserved for merge in vela CLI")
		},
	}
	rootCmd.PersistentFlags().StringVarP(&flag.ConfigFile, "config", "c", "", "set configuration file")
	rootCmd.PersistentFlags().BoolVar(&flag.Debug, "debug", false, "print debug logs")
	rootCmd.AddCommand(
		CmdCreate(&cmdConfig),
		CmdDelete(&cmdConfig),
	)

	return &rootCmd
}

func Execute() {
	cmd := NewCmdMVela()
	err := cmd.ParseFlags(os.Args)
	if err != nil {
		klog.ErrorS(err, "parse fail")
		return
	}
	err = cmd.Execute()
	if err != nil {
		klog.ErrorS(err, "execute fail")
	}
}
