package pkg

type Config struct {
	ApiVersion     string           `json:"apiVersion"`
	Kind           string           `json:"kind"`
	ManagedCluster int              `json:"managedCluster"`
	KubeconfigOpts KubeconfigOption `json:"kubeconfigOpts"`
	HelmOpts       HelmOpts         `json:"helmOpts"`
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
