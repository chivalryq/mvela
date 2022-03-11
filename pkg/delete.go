package pkg

import (
	"strings"

	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3d "github.com/rancher/k3d/v5/pkg/types"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

func CmdDelete(cmdConfig *Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "delete",
		Short: "Delete all-in-one vela environment",
		Long:  "Delete all-in-one vela environment",
		Run: func(cmd *cobra.Command, args []string) {
			clusterList, err := k3dClient.ClusterList(cmd.Context(), runtimes.SelectedRuntime)
			if err != nil {
				klog.ErrorS(err, "Fail to list clusters")
				return
			}

			if len(clusterList) == 0 {
				klog.Error("No clusters to delete, run `mvela create` first")
			}

			mvelaClusters := []*k3d.Cluster{}
			for _, c := range clusterList {
				if isMvelaCluster(c.Name) {
					mvelaClusters = append(mvelaClusters, c)
				}
			}

			// check cluster existence
			for _, r := range mvelaClusters {
				err = k3dClient.ClusterDelete(cmd.Context(), runtimes.SelectedRuntime, r, k3d.ClusterDeleteOpts{
					SkipRegistryCheck: false,
				})
				if err != nil {
					klog.ErrorS(err, "Fail to delete cluster")
					return
				}
				klog.Infof("Successfully delete cluster: %s", r.Name)

				// delete Kubeconfig

			}
		},
	}
	return &cmd
}

func isMvelaCluster(name string) bool {
	return strings.Contains(name, "mvela-cluster")
}
