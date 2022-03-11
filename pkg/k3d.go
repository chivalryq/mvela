package pkg

import (
	"errors"
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

func getClusterCreateOpts(r Registry) k3d.ClusterCreateOpts {
	InfoMirrors(r)
	k3sRegistry := convertRegistry(r)
	clusterCreateOpts := k3d.ClusterCreateOpts{
		GlobalLabels: map[string]string{}, // empty init
		GlobalEnv:    []string{},          // empty init
		Registries: registry{
			Config: &k3sRegistry,
		},
	}

	// ensure, that we have the default object labels
	for k, v := range k3d.DefaultRuntimeLabels {
		clusterCreateOpts.GlobalLabels[k] = v
	}

	return clusterCreateOpts
}

// getClusterConfig will get different k3d.Cluster based on ordinal , storage for external storage, token is needed if storage is set
func getClusterConfig(ordinal int, storage Storage, token string) (k3d.Cluster, error) {
	if storage.Endpoint != "" && token == "" {
		return k3d.Cluster{}, errors.New("token is needed if using external storage")
	}
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

	// use external storage in control plane if set
	if isControlPlane(ordinal) {
		serverNode.Args = convertStorageToNodeArgs(storage, token)
	}
	clusterConfig.Nodes = append(clusterConfig.Nodes, &serverNode)

	return clusterConfig, nil
}

func InfoMirrors(registry Registry) {
	for k, e := range registry.Mirrors {
		klog.Infof("Using registries %s -> %v\n", k, e)
	}
}

func convertStorageToNodeArgs(storage Storage, token string) []string {
	res := []string{}
	res = append(res, "--token="+token)
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

func convertRegistry(r Registry) k3s.Registry {
	kr := k3s.Registry{
		Mirrors: map[string]k3s.Mirror{},
		Configs: map[string]k3s.RegistryConfig{},
		Auths:   map[string]k3s.AuthConfig{},
	}
	for k, v := range r.Mirrors {
		k3sMirror := k3s.Mirror{
			Endpoints: v.Endpoint,
		}
		kr.Mirrors[k] = k3sMirror
	}
	for k, v := range r.Configs {
		k3sConfig := k3s.RegistryConfig{
			Auth: (*k3s.AuthConfig)(v.Auth),
			TLS:  (*k3s.TLSConfig)(v.TLS),
		}
		kr.Configs[k] = k3sConfig
	}
	for k, v := range r.Auths {
		k3sAuth := k3s.AuthConfig{
			Username:      v.Username,
			Password:      v.Password,
			Auth:          v.Auth,
			IdentityToken: v.IdentityToken,
		}
		kr.Auths[k] = k3sAuth
	}

	return kr
}
