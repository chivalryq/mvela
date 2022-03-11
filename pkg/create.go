package pkg

import (
	"context"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/kyokomi/emoji/v2"
	k3dClient "github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	"github.com/rancher/k3d/v5/pkg/runtimes"
	k3dTypes "github.com/rancher/k3d/v5/pkg/types"
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
			// create k3d
			runConfigs, err := GetClusterRunConfig(*cmdConfig)
			if err != nil {
				klog.ErrorS(err, "Fail to get cluster-run configs")
			}

			// Check cluster existence and create all cluster based on flag
			klog.Infof("Making sure directory exists %s\n", cmdConfig.KubeconfigOpts.Output)
			err = os.MkdirAll(cmdConfig.KubeconfigOpts.Output, 0o755)
			if err != nil {
				klog.ErrorS(err, "Fail to create directory to save kubeconfig")
			}

			// check cluster existence
			for ord, r := range runConfigs {
				klog.Infof("Creating Cluster No.%d: %s", ord, r.Cluster.Name)
				RunClusterIfNotExist(cmd.Context(), r)
				// kubeconfig
				KubeConfigOutput := path.Join(cmdConfig.KubeconfigOpts.Output, r.Cluster.Name)
				WriteKubeConfig(cmd.Context(), KubeConfigOutput, r.Cluster)

				// Update KUBECONFIG if control plane
				if isControlPlane(ord) {
					if cmdConfig.KubeconfigOpts.UpdateEnvironment {
						klog.Info("Setting KUBECONFIG to " + KubeConfigOutput)
						err = os.Setenv("KUBECONFIG", KubeConfigOutput)
						if err != nil {
							klog.ErrorS(err, "Fail to set environment var KUBECONFIG")
						}

						controlPlaneKubeConf = KubeConfigOutput
					}
					// install helm chart
					err = InstallVelaCore(cmdConfig.HelmOpts)
					if err != nil {
						klog.ErrorS(err, "Fail to Install helm chart, you can install manually later")
					} else {
						klog.Info("Successfully installed vela-core helm chart")
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
		if c.Name == fmt.Sprintf("k3d-%s-server-0", clusterName) {
			containerIP = strings.TrimSuffix(c.IPv4Address, "/16")
		}
	}
	kubeConfig := string(fb)
	re := regexp.MustCompile(`0.0.0.0:\d{4}`)
	internalKubeConfig := re.ReplaceAllString(kubeConfig, fmt.Sprintf("%s:6443", containerIP))

	err = os.WriteFile(fmt.Sprintf("%s-internal", kubeconfigFile), []byte(internalKubeConfig), 0o600)
	if err != nil {
		klog.ErrorS(err, "Fail to write internal kubeconfig", "cluster", clusterName)
		return err
	}
	return nil
}

func printGuide(cfg Config) {
	fmt.Println()
	emoji.Fprintln(os.Stdout, ":rocket: Successfully setup KubeVela control plane (and subClusters)")
	if cfg.KubeconfigOpts.UpdateEnvironment {
		emoji.Fprintf(os.Stdout, ":pushpin: Have set KUBECONFIG=%s\n", controlPlaneKubeConf)
	} else {
		emoji.Fprintf(os.Stdout, ":pushpin: Set KUBECONFIG=%s to connect to cluster\n", controlPlaneKubeConf)
	}

	emoji.Fprintf(os.Stdout, ":telescope: See usable components, run `vela components`\n")
	if cfg.ManagedCluster > 1 {
		internalCfg := path.Join(cfg.KubeconfigOpts.Output, "mvela-cluster-1-internal")
		subCfg := path.Join(cfg.KubeconfigOpts.Output, "mvela-cluster-1")
		emoji.Fprintf(os.Stdout, ":link: Join sub-clusters, run `vela cluster join %s`, or more with other number\n", internalCfg)
		emoji.Fprintf(os.Stdout, ":key: Check sub-clusters, run `KUBECONFIG=%s kubectl get pod -A`, or more with other number\n", subCfg)
	}
}
