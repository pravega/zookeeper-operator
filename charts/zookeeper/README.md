# Zookeeper Helm Chart

Installs Zookeeper clusters atop Kubernetes.

## Introduction

This chart creates a Zookeeper cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Zookeeper cluster in multiple namespaces.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3+
  - Zookeeper Operator. You can install it using its own [Helm chart](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper-operator)

## Installing the Chart

To install the chart with the release name `my-release`:

```
$ helm install my-release zookeeper
```

The command deploys zookeeper on the Kubernetes cluster in the default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```
$ helm uninstall my-release
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the Zookeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `replicas` | Expected size of the zookeeper cluster (valid range is from 1 to 7) | `3` |
| `image.repository` | Image repository | `pravega/zookeeper` |
| `image.tag` | Image tag | `0.2.8` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `domainName` | Domain name to be used for DNS | |
| `labels` | Specifies the labels to be attached | `{}` |
| `ports` | Groups the ports for a zookeeper cluster node for easy access | `[]` |
| `pod` | Defines the policy to create new pods for the zookeeper cluster | `{}` |
| `pod.labels` | Labels to attach to the pods | `{}` |
| `pod.nodeSelector` | Map of key-value pairs to be present as labels in the node in which the pod should run | `{}` |
| `pod.affinity` | Specifies scheduling constraints on pods | `{}` |
| `pod.resources` | Specifies resource requirements for the container | `{}` |
| `pod.tolerations` | Specifies the pod's tolerations | `[]` |
| `pod.env` | List of environment variables to set in the container | `[]` |
| `pod.annotations` | Specifies the annotations to attach to pods | `{}` |
| `pod.securityContext` | Specifies the security context for the entire pod | |
| `pod.terminationGracePeriodSeconds` | Amount of time given to the pod to shutdown normally | `180` |
| `config.initLimit` | Amount of time (in ticks) to allow followers to connect and sync to a leader | `10` |
| `config.tickTime` | Length of a single tick which is the basic time unit used by Zookeeper (measured in milliseconds) | `2000` |
| `config.syncLimit` | Amount of time (in ticks) to allow followers to sync with Zookeeper | `2` |
| `config.quorumListenOnAllIPs` | Whether Zookeeper server will listen for connections from its peers on all available IP addresses | `false` |
| `persistence.reclaimPolicy` | Reclaim policy for persistent volumes | `Delete` |
| `persistence.storageClassName` | Storage class for persistent volumes | `standard` |
| `persistence.volumeSize` | Size of the volume requested for persistent volumes | `20Gi` |
