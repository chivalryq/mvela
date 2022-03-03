package pkg

import (
	"github.com/rancher/k3d/v5/pkg/types/k3s"
)

type Config struct {
	ApiVersion     string           `json:"apiVersion"`
	Kind           string           `json:"kind"`
	ManagedCluster int              `json:"managedCluster"`
	KubeconfigOpts KubeconfigOption `json:"kubeconfigOpts"`
	HelmOpts       HelmOpts         `json:"helmOpts"`
	Registries     k3s.Registry     `json:"registries"`
}

type KubeconfigOption struct {
	Output            string `json:"output"`
	UpdateEnvironment bool   `json:"updateEnvironment"`
}

type HelmOpts struct {
	Type      string `json:"type"`
	ChartPath string `json:"chartPath"`
	Version   string `json:"version"`
}
