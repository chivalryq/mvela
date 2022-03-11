# mvela

a CLI tool for [KubeVela](https://github.com/oam-dev/kubevela) trial

![Demo](./example/demo.gif) 

## Prerequisites

 - docker
 - vela CLI

## Run

1. `make`
2. `bin/mvela create`

## Clean up

`make uninstall`

## Configuration

Add following snippets to config file. Run it with `mvela create -c conf.yaml`

```yaml
apiVersion: mvela.oam.dev/v1alpha1
kind: Simple
managedCluster: 2 # cluster numbers, 1st cluster will be seen as control plane
kubeconfigOpts:	
  output: /Users/qiaozp/.vela/kubeConfig # directory to write KubeConfigs
  updateEnvironment: true # whether update KUBECONFIG var in your shell
```


## Known Issue

1. `make uninstall` will stop the container, but not clean up k3d data. In another word it's possible to restart the cluster by k3d CLI. That's not complete uninstall
