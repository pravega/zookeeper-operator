# Zookeeper Operator Helm Chart

Installs [Zookeeper Operator](https://github.com/pravega/zookeeper-operator) to create/configure/manage Zookeeper clusters atop Kubernetes.

## Introduction

This chart bootstraps a [Zookeeper Operator](https://github.com/pravega/zookeeper-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install my-release pravega/zookeeper-operator --version=`version`
```
Note: `version` can be any stable release version for zookeeper operator from 0.2.8 onwards.

The above command deploys zookeeper-operator on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the deployment `my-release`:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Zookeeper operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/zookeeper-operator` |
| `image.tag` | Image tag | `0.2.8` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `crd.create` | Create zookeeper CRD | `true` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Name for the service account | `zookeeper-operator` |
| `watchNamespace` | Namespaces to be watched  | `""` |
| `resources` | Specifies resource requirements for the container | `{}` |
| `nodeSelector` | Map of key-value pairs to be present as labels in the node in which the pod should run | `{}` |
| `affinity` | Specifies scheduling constraints on pods | `{}` |
| `tolerations` | Specifies the pod's tolerations | `[]` |
