package pkg

import (
	_ "embed"
	"errors"
	"os"
	"path"

	config "github.com/rancher/k3d/v5/pkg/config/v1alpha4"
	"github.com/spf13/viper"
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
	viper.SetConfigFile(ConfigFile)
	err := bindEnv()
	if err != nil {
		klog.Error("Fail to bind environment to mvela config")
	}
	err = viper.ReadInConfig()

	if err != nil {
		klog.ErrorS(err, "Fail to read config file")
	}
	err = viper.Unmarshal(&res)
	if err != nil {
		klog.ErrorS(err, "Fail to unmarshal config file, gonna use default configs")
		return initDefaultConfig()
	}
	reportConf(res)

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
		cluster := getClusterConfig(ord, cmdConfig.Storage)
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

func bindEnv() error {
	if err := viper.BindEnv("storage.endpoint", "DATASTORE_ENDPOINT"); err != nil {
		return err
	}
	if err := viper.BindEnv("storage.ca_file", "DATASTORE_CAFILE"); err != nil {
		return err
	}
	if err := viper.BindEnv("storage.key_file", "DATASTORE_KEYFILE"); err != nil {
		return err
	}
	return nil
}

func reportConf(c Config) {
	klog.Info("Gonna use configuration")
	klog.Info(c)
}
