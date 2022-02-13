package pkg

import (
	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	l "github.com/rancher/k3d/v5/pkg/logger"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3d "github.com/rancher/k3d/v5/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func CmdDelete(cmdConfig *Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete",
		Short: "Delete all-in-one vela environment",
		Long:  "Delete all-in-one vela environment",
		Run: func(cmd *cobra.Command, args []string) {
			l.Log().SetLevel(logrus.FatalLevel)

			// create k3d
			runConfig := GetClusterRunConfig()

			// check cluster existence
			if _, err = k3dClient.ClusterGet(cmd.Context(), runtimes.SelectedRuntime, &runConfig.Cluster); err != nil {
				klog.Fatal("Fail to delete cluster because it doesn't exist")
				return
			}

			err = k3dClient.ClusterDelete(cmd.Context(), runtimes.SelectedRuntime, &runConfig.Cluster,k3d.ClusterDeleteOpts{
				SkipRegistryCheck: false,
			})
			if err != nil {
				klog.ErrorS(err, "Fail to delete cluster")
				return
			}
			klog.Info("Successfully delete cluster")
		},
	}
	return &cmd
}