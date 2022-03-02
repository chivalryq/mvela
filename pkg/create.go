package pkg

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/kyokomi/emoji/v2"
	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	l "github.com/rancher/k3d/v5/pkg/logger"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3dTypes "github.com/rancher/k3d/v5/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
)

var (
	DefaultConfigFile    = "./example/conf.yaml"
	dockerCli            client.APIClient
	controlPlaneKubeConf string
)

func init() {
	var err error
	dockerCli, err = client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}
}

func CmdCreate(cmdConfig *Config) *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "Create a all-in-one vela environment",
		Long:  "Create a all-in-one vela image and run it",
		Run: func(cmd *cobra.Command, args []string) {
			l.Log().SetLevel(logrus.FatalLevel)

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
			printGuide(*cmdConfig)
		},
	}
	return &cmd
}

func isControlPlane(ord int) bool {
	return ord == 0
}

func RunClusterIfNotExist(ctx context.Context, cluster config.ClusterConfig) {
	if _, err := k3dClient.ClusterGet(ctx, runtimes.SelectedRuntime, &cluster.Cluster); err == nil {
		klog.Infof("Detect an existing cluster: %s", cluster.Cluster.Name)
		return
	}
	err := k3dClient.ClusterRun(ctx, runtimes.SelectedRuntime, &cluster)
	if err != nil {
		klog.ErrorS(err, "Fail to create cluster", "cluster-name", cluster.Cluster.Name)
		return
	}
	klog.Infof("Successfully create cluster: %s", cluster.Cluster.Name)
}

// WriteKubeConfig write kubeconfig to output.
// There are two kinds of kubeconfig:
// mvela-cluster-n for accessing cluster from host. mvela-cluster-n-internal for accessing between clusters
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

	if !strings.Contains(cluster.Name, "control-plane") {
		err = generateInternal(ctx, output, cluster.Name)
		if err != nil {
			klog.Error("Fail to write internal kubeconfig, unable to use vela join now")
		}
	}
}

func generateInternal(ctx context.Context, kubeconfigFile string, clusterName string) error {
	klog.Info("Generating kubeconfig files for inter-cluster accessibility")
	fb, err := os.ReadFile(kubeconfigFile)
	if err != nil {
		klog.ErrorS(err, "Fail to read kubeconfig file")
	}
	// find cluster name

	networks, err := dockerCli.NetworkInspect(ctx, "k3d-mvela", types.NetworkInspectOptions{})
	if err != nil {
		klog.ErrorS(err, "Fail to inspect docker network")
		return err
	}
	var containerIP string
	cs := networks.Containers
	for _, c := range cs {
		if c.Name == fmt.Sprintf("%s-server-0", clusterName) {
			containerIP = strings.TrimSuffix(c.IPv4Address, "/16")
		}
	}
	kubeConfig := string(fb)
	internalKubeConfig := strings.Replace(kubeConfig, "0.0.0.0", containerIP, 1)
	err = os.WriteFile(fmt.Sprintf("%s-internal", kubeconfigFile), []byte(internalKubeConfig), 0o600)
	if err != nil {
		klog.ErrorS(err, "Fail to write internal kubeconfig", "cluster", clusterName)
		return err
	}
	return nil
}

func printGuide(cfg Config) {
	if cfg.KubeconfigOpts.UpdateEnvironment {
		klog.Infof("Have set KUBECONFIG=%s\n", controlPlaneKubeConf)
	} else {
		klog.Infof("Set KUBECONFIG=%s\n to connect to cluster", controlPlaneKubeConf)
	}

	emoji.Fprintf(os.Stdout, ":magnifying glass tilted left: See usable components, run `vela components`\n")
	if cfg.ManagedCluster > 1 {
		internalCfg := path.Join(cfg.KubeconfigOpts.Output, "mvela-cluster-1-internal")
		emoji.Fprintln(os.Stdout, "link: Join sub-clusters, run `vela cluster join %s`, or more with other number\n", internalCfg)
	}
}
