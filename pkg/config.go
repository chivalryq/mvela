package pkg

import (
	_ "embed"
	"fmt"
	"os"
	"path"

	"github.com/spf13/viper"

	"github.com/docker/go-connections/nat"
	"github.com/rancher/k3d/v5/pkg/client"
	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	k3d "github.com/rancher/k3d/v5/pkg/types"
	"k8s.io/klog/v2"
)

const (
	configName string = "mvela"
	k3dPrefix  string = "k3d"
)

var (
	// VelaDir is ~/.vela
	velaDir string
)

func VelaDir() string {
	if velaDir != "" {
		return velaDir
	}
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	velaDir = path.Join(homeDir, ".vela")
	return velaDir
}

func initDefaultConfig() (Config, error) {
	velaDir := VelaDir()
	return Config{
		ApiVersion:     "mvela.oam.dev/v1alpha1",
		Kind:           "Simple",
		ManagedCluster: 1,
		KubeconfigOpts: KubeconfigOption{
			Output:            path.Join(velaDir, "config", "k3d-kubeconfig"),
			UpdateEnvironment: true,
		},
	}, nil
}

func ReadConfig(ConfigFile string) (Config, error) {
	res := Config{}
	if ConfigFile == "" {
		return initDefaultConfig()
	}
	var viperCfg viper.Viper
	err := viperCfg.Unmarshal(&res)
	if err != nil {
		return Config{}, err
	}
	return res, nil
}

func getClusterCreateOpts() k3d.ClusterCreateOpts {
	clusterCreateOpts := k3d.ClusterCreateOpts{
		GlobalLabels: map[string]string{}, // empty init
		GlobalEnv:    []string{},          // empty init
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}
	return clusterCreateOpts
}

func getKubeconfigOptions() config.SimpleConfigOptionsKubeconfig {
	opts := config.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: true,
		SwitchCurrentContext:    true,
	}
	return opts
}

func GetClusterRunConfig() config.ClusterConfig {
	cluster := getClusterConfig()
	createOpts := getClusterCreateOpts()
	kubeconfigOpts := getKubeconfigOptions()

	return config.ClusterConfig{
		Cluster:           cluster,
		ClusterCreateOpts: createOpts,
		KubeconfigOpts:    kubeconfigOpts,
	}
}

func getClusterConfig() k3d.Cluster {
	// network
	k3dNetwork := k3d.ClusterNetwork{
		Name:     fmt.Sprintf("%s-%s", k3dPrefix, configName),
		External: false,
	}

	// api
	kubeAPIExposureOpts := k3d.ExposureOpts{
		Host: k3d.DefaultAPIHost,
	}
	kubeAPIExposureOpts.Port = k3d.DefaultAPIPort
	kubeAPIExposureOpts.Binding = nat.PortBinding{
		HostIP:   k3d.DefaultAPIHost,
		HostPort: "6443",
	}

	// fill cluster config
	clusterConfig := k3d.Cluster{
		Name:    "mvela-cluster",
		Network: k3dNetwork,
		KubeAPI: &kubeAPIExposureOpts,
	}

	klog.Info("disabling load balancer")

	// nodes
	clusterConfig.Nodes = []*k3d.Node{}

	serverNode := k3d.Node{
		Name:       client.GenerateNodeName(clusterConfig.Name, k3d.ServerRole, 0),
		Role:       k3d.ServerRole,
		Image:      "rancher/k3s:latest",
		ServerOpts: k3d.ServerOpts{},
	}
	clusterConfig.Nodes = append(clusterConfig.Nodes, &serverNode)

	// opts
	return clusterConfig

}
