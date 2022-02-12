package pkg

import (
	"fmt"
	"os"

	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	l "github.com/rancher/k3d/v5/pkg/logger"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	DefaultConfigFile = "./example/conf.yaml"
)

func CmdCreate(cmdConfig *Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "Create a all-in-one vela environment",
		Long:  "Create a all-in-one vela image and run it",
		Run: func(cmd *cobra.Command, args []string) {
			l.Log().SetLevel(logrus.FatalLevel)

			// create k3d
			runConfig := GetClusterRunConfig()
			err = k3dClient.ClusterRun(cmd.Context(), runtimes.SelectedRuntime, &runConfig)
			if err != nil {
				klog.ErrorS(err, "create cluster error")
				return
			}
			klog.Info("Successfully create cluster")

			// kubeconfig
			kubeconfigOpt := k3dClient.WriteKubeConfigOptions{UpdateExisting: false, OverwriteExisting: true, UpdateCurrentContext: false}
			if _, err = k3dClient.KubeconfigGetWrite(cmd.Context(), runtimes.SelectedRuntime, &runConfig.Cluster, cmdConfig.KubeconfigOpts.Output, &kubeconfigOpt); err != nil {
				klog.ErrorS(err, "fail to write kubeconfig")
			}
			klog.Info("Successfully generate kubeconfig file at ", cmdConfig.KubeconfigOpts.Output)
			if cmdConfig.KubeconfigOpts.UpdateEnvironment {
				_ = os.Setenv("KUBECONFIG", cmdConfig.KubeconfigOpts.Output)
			}

			// install helm chart
			err = InstallVelaCore(cmdConfig.HelmOpts)
			if err != nil {
				klog.ErrorS(err, "Fail to Install helm chart, you can install manually later")
			} else {
				fmt.Println("Successfully installed release kubevela")
			}

			// feedback
			if cmdConfig.KubeconfigOpts.UpdateEnvironment {
				klog.Infof("Have set KUBECONFIG=%s\n", cmdConfig.KubeconfigOpts.Output)
			} else {
				klog.Infof("Set KUBECONFIG=%s\n to connect to cluster", cmdConfig.KubeconfigOpts.Output)
			}
			fmt.Println("run `vela components`")
		},
	}
	return &cmd
}
