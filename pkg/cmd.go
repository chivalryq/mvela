package pkg

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

const VelaCoreVersion = "1.2.2"

type rootFlag struct {
	DebugLog   bool
	ConfigFile string
}

var (
	flag      rootFlag
	cmdConfig Config
	err       error
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
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("reserved for merge in vela CLI")
		},
	}
	rootCmd.PersistentFlags().StringVarP(&flag.ConfigFile, "config", "c", "", "set configuration file")
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
