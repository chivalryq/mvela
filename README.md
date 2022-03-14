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

#### Run with external database

**Note: `token` field or TOKEN environment variable is needed when using external database for connect to DB repeatably** 

When you want keep whole cluster metadata in an external DB, you can create a cluster based on an external DB. You can specify the storage field of configuration like

```yaml
apiVersion: mvela.oam.dev/v1alpha1
kind: Simple
# other fields..

storage:
  endpoint: mysql://tcp(user:passwd@host:PORT)/DBNAME
  cert_file: /path/to/client.crt
  key_file: /path/to/client.key
token: SECRET
```

Keep database connection string in shell is more recommended. you can run like:

```shell
DATASTORE_ENDPOINT=mysql://tcp(user:passwd@host:PORT)/DBNAME TOKEN=SECRET mvela create
```
| field in config   | environment var    |
|--------------- | --------------- |
| storage.endpoint   | DATASTORE_ENDPOINT   |
| storage.cert_file   | DATASTORE_CAFILE   |
| storage   | DATASTORE_KEYFILE   |

For the connection string format, See k3s [doc](https://rancher.com/docs/k3s/latest/en/installation/datastore/#datastore-endpoint-format-and-functionality) 

## Known Issue

1. `make uninstall` will stop the container, but not clean up k3d data. In another word it's possible to restart the cluster by k3d CLI. That's not complete uninstall
