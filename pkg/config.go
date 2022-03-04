package pkg

import (
	_ "embed"
	"errors"
	"os"
	"path"

	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

const (
	configName string = "mvela"
	k3dPrefix  string = "k3d"
)

// VelaDir is ~/.vela
var velaDir string

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
			Output:            path.Join(velaDir, "kubeConfig"),
			UpdateEnvironment: true,
		},
	}, nil
}

func ReadConfig(ConfigFile string) (Config, error) {
	res := Config{}
	if ConfigFile == "" {
		_, err := os.Stat("example/conf.yaml")
		if err != nil && errors.Is(err, os.ErrNotExist) {
			return initDefaultConfig()
		}
		ConfigFile = "example/conf.yaml"
		klog.Infof("Using config file: %s\n", ConfigFile)
	}
	b, err := os.ReadFile(ConfigFile)
	if err != nil {
		klog.ErrorS(err, "Fail to read config file, gonna use default configs")
		return initDefaultConfig()
	}
	// viper.SetConfigFile(ConfigFile)
	// viper.ReadInConfig()
	err = yaml.Unmarshal(b, &res)
	// err := viper.GetViper().Unmarshal(&res)
	if err != nil {
		klog.ErrorS(err, "Fail to unmarshal config file, gonna use default configs")
		return initDefaultConfig()
	}

	return CompleteConfig(res), nil
}

// CompleteConfig validate and complete the config
func CompleteConfig(origin Config) Config {
	complete := origin
	if origin.ManagedCluster < 1 {
		klog.Infof("Invalid configuration for managedCluster field: %d, set to 1", origin.ManagedCluster)
		complete.ManagedCluster = 1
	}
	return complete
}

func getKubeconfigOptions() config.SimpleConfigOptionsKubeconfig {
	opts := config.SimpleConfigOptionsKubeconfig{
		UpdateDefaultKubeconfig: true,
		SwitchCurrentContext:    true,
	}
	return opts
}

func GetClusterRunConfig(cmdConfig Config) []config.ClusterConfig {
	managedCluster := cmdConfig.ManagedCluster
	runConfigs := []config.ClusterConfig{}
	for ord := 0; ord < managedCluster; ord++ {
		cluster := getClusterConfig(ord)
		createOpts := getClusterCreateOpts(cmdConfig.Registries)
		kubeconfigOpts := getKubeconfigOptions()
		runConfigs = append(runConfigs, config.ClusterConfig{
			Cluster:           cluster,
			ClusterCreateOpts: createOpts,
			KubeconfigOpts:    kubeconfigOpts,
		})
	}
	return runConfigs
}
