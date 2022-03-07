package pkg

import (
	"fmt"

	"github.com/docker/go-connections/nat"
	"github.com/rancher/k3d/v5/pkg/client"
	k3d "github.com/rancher/k3d/v5/pkg/types"
	"github.com/rancher/k3d/v5/pkg/types/k3s"
	"k8s.io/klog/v2"
)

type registry struct {
	Create *k3d.Registry   `yaml:"create,omitempty" json:"create,omitempty"`
	Use    []*k3d.Registry `yaml:"use,omitempty" json:"use,omitempty"`
	Config *k3s.Registry   `yaml:"config,omitempty" json:"config,omitempty"`
}

func getClusterCreateOpts(r k3s.Registry) k3d.ClusterCreateOpts {
	InfoMirrors(r)
	clusterCreateOpts := k3d.ClusterCreateOpts{
		GlobalLabels: map[string]string{}, // empty init
		GlobalEnv:    []string{},          // empty init
		Registries: registry{
			Config: &r,
		},
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}

	return clusterCreateOpts
}

// getClusterConfig will get different k3d.Cluster based on ordinal and storage arguments
func getClusterConfig(ordinal int, storage Storage) k3d.Cluster {
	// All cluster will be created in one docker network
	universalK3dNetwork := k3d.ClusterNetwork{
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
		HostPort: fmt.Sprint(6443 + ordinal),
	}

	// fill cluster config
	var clusterName string
	if ordinal == 0 {
		clusterName = "mvela-cluster-control-plane"
	} else {
		clusterName = fmt.Sprintf("mvela-cluster-%d", ordinal)
	}
	clusterConfig := k3d.Cluster{
		Name:    clusterName,
		Network: universalK3dNetwork,
		KubeAPI: &kubeAPIExposureOpts,
	}

	// klog.Info("disabling load balancer")

	// nodes
	clusterConfig.Nodes = []*k3d.Node{}

	serverNode := k3d.Node{
		Name:       client.GenerateNodeName(clusterConfig.Name, k3d.ServerRole, 0),
		Role:       k3d.ServerRole,
		Image:      "rancher/k3s:latest",
		ServerOpts: k3d.ServerOpts{},
	}

	// use external storage if set
	if ordinal == 0 {
		serverNode.Args = convertStorageToNodeArgs(storage)
	}
	clusterConfig.Nodes = append(clusterConfig.Nodes, &serverNode)

	return clusterConfig
}

func InfoMirrors(registry k3s.Registry) {
	for k, e := range registry.Mirrors {
		klog.Infof("Using registries %s -> %v\n", k, e)
	}
}

func convertStorageToNodeArgs(storage Storage) []string {
	res := []string{}
	if storage.Endpoint != "" {
		res = append(res, "--datastore-endpoint="+storage.Endpoint)
	}
	if storage.CAFile != "" {
		res = append(res, "--datastore-cafile="+storage.CAFile)
	}
	if storage.CertFile != "" {
		res = append(res, "--datastore-certfile="+storage.CertFile)
	}
	if storage.KeyFile != "" {
		res = append(res, "--datastore-keyfile="+storage.KeyFile)
	}
	return res
}
