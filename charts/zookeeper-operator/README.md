# Zookeeper Operator Helm Chart

Installs [Zookeeper Operator](https://github.com/pravega/zookeeper-operator) to create/configure/manage Zookeeper clusters atop Kubernetes.

## Introduction

This chart bootstraps a [Zookeeper Operator](https://github.com/pravega/zookeeper-operator) deployment on a [Kubernetes](http://kubernetes.io) cluster using the [Helm](https://helm.sh) package manager.

## Prerequisites
  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+

## Installing the Chart

To install the zookeeper-operator chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/zookeeper-operator --version=[VERSION]
```
- **[RELEASE_NAME]** is the release name for the zookeeper-operator chart.
- **[DEPLOYMENT_NAME]** is the name of the zookeeper-operator deployment so created. (If [RELEASE_NAME] contains the string `zookeeper-operator`, `[DEPLOYMENT_NAME] = [RELEASE_NAME]`, else `[DEPLOYMENT_NAME] = [RELEASE_NAME]-zookeeper-operator`. The [DEPLOYMENT_NAME] can however be overridden by providing `--set fullnameOverride=[DEPLOYMENT_NAME]` along with the helm install command)
- **[VERSION]** can be any stable release version for zookeeper-operator from 0.2.8 onwards.

This command deploys a zookeeper-operator on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the zookeeper-operator chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the zookeeper-operator chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `image.repository` | Image repository | `pravega/zookeeper-operator` |
| `image.tag` | Image tag | `0.2.12` |
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
| `additionalEnv` | Additional Environment Variables | `[]` |
| `additionalSidecars` | Additional Sidecars Configuration | `[]` |
| `additionalVolumes` | Additional volumes required for sidecars | `[]` |
