apiVersion:     "mvela.oam.dev/v1alpha1"
kind:           "Simple"
managedCluster: *0 | int & >=0
kubeconfigOpts: {
	output:            *"~/.vela/config/mvela.yaml" | string
	updateEnvironment: *true | bool
}
helmOpts: {
	type:      *"helm" | "local"
	chartPath: *"" | string
	version:   string
}
