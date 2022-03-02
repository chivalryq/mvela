package pkg

import (
	"context"
	"fmt"
	"os"
	"path"

	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	l "github.com/rancher/k3d/v5/pkg/logger"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3dTypes "github.com/rancher/k3d/v5/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var DefaultConfigFile = "./example/conf.yaml"

func CmdCreate(cmdConfig *Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "Create a all-in-one vela environment",
		Long:  "Create a all-in-one vela image and run it",
		Run: func(cmd *cobra.Command, args []string) {
			l.Log().SetLevel(logrus.FatalLevel)
			var controlPlaneKubeConf string

			// create k3d
			runConfigs := GetClusterRunConfig(cmdConfig.ManagedCluster)

			// Check cluster existence and create all cluster based on flag
			err := os.MkdirAll(cmdConfig.KubeconfigOpts.Output, 0o755)
			if err != nil {
				klog.ErrorS(err, "Fail to create directory to save kubeconfig")
			}

			// check cluster existence
			for ord, r := range runConfigs {
				RunClusterIfNotExist(cmd.Context(), r)
				// kubeconfig
				KubeConfigOutput := path.Join(cmdConfig.KubeconfigOpts.Output, r.Cluster.Name)
				WriteKubeConfig(cmd.Context(), KubeConfigOutput, r.Cluster)

				// Update KUBECONFIG if control plane
				if isControlPlane(ord) {
					if cmdConfig.KubeconfigOpts.UpdateEnvironment {
						_ = os.Setenv("KUBECONFIG", KubeConfigOutput)
						controlPlaneKubeConf = KubeConfigOutput
					}
					// install helm chart
					err = InstallVelaCore(cmdConfig.HelmOpts)
					if err != nil {
						klog.ErrorS(err, "Fail to Install helm chart, you can install manually later")
					} else {
						fmt.Println("Successfully installed vela-core helm chart")
					}
				}
			}

			// feedback
			if cmdConfig.KubeconfigOpts.UpdateEnvironment {
				klog.Infof("Have set KUBECONFIG=%s\n", controlPlaneKubeConf)
			} else {
				klog.Infof("Set KUBECONFIG=%s\n to connect to cluster", controlPlaneKubeConf)
			}
			fmt.Println("run `vela components`")
		},
	}
	return &cmd
}

func isControlPlane(ord int) bool {
	return ord == 0
}

func RunClusterIfNotExist(ctx context.Context, cluster config.ClusterConfig) {
	if _, err = k3dClient.ClusterGet(ctx, runtimes.SelectedRuntime, &cluster.Cluster); err == nil {
		klog.Infof("Detect an existing cluster: %s", cluster.Cluster.Name)
		return
	}
	err = k3dClient.ClusterRun(ctx, runtimes.SelectedRuntime, &cluster)
	if err != nil {
		klog.ErrorS(err, "Fail to create cluster", "cluster-name", cluster.Cluster.Name)
		return
	}
	klog.Infof("Successfully create cluster: %s", cluster.Cluster.Name)
}

// WriteKubeConfig write kubeconfig to output
func WriteKubeConfig(ctx context.Context, output string, cluster k3dTypes.Cluster) {
	_, err := os.Stat(output)
	if err == nil {
		klog.Infof("Overwriting the mvela kubeconfig at %s", output)
	}
	kubeconfigOpt := k3dClient.WriteKubeConfigOptions{UpdateExisting: false, OverwriteExisting: true, UpdateCurrentContext: false}
	if _, err = k3dClient.KubeconfigGetWrite(ctx, runtimes.SelectedRuntime, &cluster, output, &kubeconfigOpt); err == nil {
		klog.Info("Successfully generate kubeconfig file at ", output)
	} else {
		klog.ErrorS(err, "Fail to write kubeconfig")
	}
}
