package pkg

import (
	"github.com/rancher/k3d/v5/pkg/types/k3s"
)

type Config struct {
	ApiVersion     string           `json:"apiVersion" yaml:"apiVersion"`
	Kind           string           `json:"kind" yaml:"kind"`
	ManagedCluster int              `json:"managedCluster" yaml:"managedCluster"`
	KubeconfigOpts KubeconfigOption `json:"kubeconfigOpts" yaml:"kubeconfigOpts"`
	HelmOpts       HelmOpts         `json:"helmOpts" yaml:"helmOpts"`
	Registries     k3s.Registry     `json:"registries" yaml:"registries"`
	Storage        Storage          `json:"storage"`
}

type KubeconfigOption struct {
	Output            string `json:"output" yaml:"output"`
	UpdateEnvironment bool   `json:"updateEnvironment" yaml:"updateEnvironment"`
}

type HelmOpts struct {
	Type      string `json:"type" yaml:"type"`
	ChartPath string `json:"chartPath" yaml:"chartPath"`
	Version   string `json:"version" yaml:"version"`
}

type Storage struct {
	Endpoint string `json:"endpoint" yaml:"endpoint"`
	CAFile   string `json:"ca_file" yaml:"ca_file"`
	CertFile string `json:"cert_file" yaml:"cert_file"`
	KeyFile  string `json:"key_file" yaml:"key_file"`
}
