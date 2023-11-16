# Zookeeper Helm Chart

Installs Zookeeper clusters atop Kubernetes.

## Introduction

This chart creates a Zookeeper cluster in [Kubernetes](http://kubernetes.io) using the [Helm](https://helm.sh) package manager. The chart can be installed multiple times to create Zookeeper cluster in multiple namespaces.

## Prerequisites

  - Kubernetes 1.15+ with Beta APIs
  - Helm 3.2.1+
  - Zookeeper Operator. You can install it using its own [Helm chart](https://github.com/pravega/zookeeper-operator/tree/master/charts/zookeeper-operator)

## Installing the Chart

To install the zookeeper chart, use the following commands:

```
$ helm repo add pravega https://charts.pravega.io
$ helm repo update
$ helm install [RELEASE_NAME] pravega/zookeeper --version=[VERSION]
```
where:
- **[RELEASE_NAME]** is the release name for the zookeeper chart.
- **[CLUSTER_NAME]** is the name of the zookeeper cluster so created. (If [RELEASE_NAME] contains the string `zookeeper`, `[CLUSTER_NAME] = [RELEASE_NAME]`, else `[CLUSTER_NAME] = [RELEASE_NAME]-zookeeper`. The [CLUSTER_NAME] can however be overridden by providing `--set fullnameOverride=[CLUSTER_NAME]` along with the helm install command)
- **[VERSION]** can be any stable release version for zookeeper from 0.2.8 onwards.

This command deploys zookeeper on the Kubernetes cluster in its default configuration. The [configuration](#configuration) section lists the parameters that can be configured during installation.

## Upgrading the Chart

To upgrade the zookeeper chart from version **[OLD_VERSION]** to version **[NEW_VERSION]**, use the following command:

```
$ helm upgrade [RELEASE_NAME] pravega/zookeeper --version=[NEW_VERSION] --set image.tag=[NEW_VERSION] --reuse-values --timeout 600s
```
**Note:** By specifying the `--reuse-values` option, the configuration of all parameters are retained across upgrades. However if some values need to be modified during the upgrade, the `--set` flag can be used to specify the new configuration for these parameters. Also, by skipping the `reuse-values` flag, the values of all parameters are reset to the default configuration that has been specified in the published charts for version [NEW_VERSION].

## Uninstalling the Chart

To uninstall/delete the zookeeper chart, use the following command:

```
$ helm uninstall [RELEASE_NAME]
```

This command removes all the Kubernetes components associated with the chart and deletes the release.

## Configuration

The following table lists the configurable parameters of the zookeeper chart and their default values.

| Parameter | Description | Default |
| ----- | ----------- | ------ |
| `replicas` | Expected size of the zookeeper cluster (valid range is from 1 to 7) | `3` |
| `maxUnavailableReplicas` | Max unavailable replicas in pdb | `1` |
| `triggerRollingRestart` | If true, the zookeeper cluster is restarted. After the restart is triggered, this value is auto-reverted to false. | `false` |
| `image.repository` | Image repository | `pravega/zookeeper` |
| `image.tag` | Image tag | `0.2.15` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `domainName` | External host name appended for dns annotation | |
| `kubernetesClusterDomain` | Domain of the kubernetes cluster | `cluster.local` |
| `probes.readiness.initialDelaySeconds` | Number of seconds after the container has started before readiness probe is initiated | `10` |
| `probes.readiness.periodSeconds` | Number of seconds in which readiness probe will be performed | `10` |
| `probes.readiness.failureThreshold` | Number of seconds after which the readiness probe times out | `3` |
| `probes.readiness.successThreshold` | Minimum number of consecutive successes for the readiness probe to be considered successful after having failed | `1` |
| `probes.readiness.timeoutSeconds` | Number of times Kubernetes will retry after a readiness probe failure before restarting the container | `10` |
| `probes.liveness.initialDelaySeconds` | Number of seconds after the container has started before liveness probe is initiated | `10` |
| `probes.liveness.periodSeconds` | Number of seconds in which liveness probe will be performed  | `10` |
| `probes.liveness.failureThreshold` | Number of seconds after which the liveness probe times out | `3` |
| `probes.liveness.timeoutSeconds` | Number of times Kubernetes will retry after a liveness probe failure before restarting the container | `10` |
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
| `pod.securityContext` | Specifies the security context for the entire pod | `{}` |
| `pod.terminationGracePeriodSeconds` | Amount of time given to the pod to shutdown normally | `30` |
| `pod.serviceAccountName` | Name for the service account | `zookeeper` |
| `pod.imagePullSecrets` | ImagePullSecrets is a list of references to secrets in the same namespace to use for pulling any images. | `[]` |
| `clientService` | Defines the policy to create client Service for the zookeeper cluster. | {} |
| `clientService.annotations` | Specifies the annotations to attach to client Service the operator creates. | {} |
| `headlessService` | Defines the policy to create headless Service for the zookeeper cluster. | {} |
| `headlessService.annotations` | Specifies the annotations to attach to headless Service the operator creates. | {} |
| `adminServerService` | Defines the policy to create AdminServer Service for the zookeeper cluster. | {} |
| `adminServerService.annotations` | Specifies the annotations to attach to AdminServer Service the operator creates. | {} |
| `adminServerService.external` | Specifies if LoadBalancer should be created for the AdminServer. True means LoadBalancer will be created, false - only ClusterIP will be used. | false |
| `config.initLimit` | Amount of time (in ticks) to allow followers to connect and sync to a leader | `10` |
| `config.tickTime` | Length of a single tick which is the basic time unit used by Zookeeper (measured in milliseconds) | `2000` |
| `config.syncLimit` | Amount of time (in ticks) to allow followers to sync with Zookeeper | `2` |
| `config.globalOutstandingLimit` | Max limit for outstanding requests | `1000` |
| `config.preAllocSize` | PreAllocSize in kilobytes | `65536` |
| `config.snapCount` | The number of transactions recorded in the transaction log before a snapshot can be taken | `100000` |
| `config.commitLogCount` | The number of committed requests in memory | `500`
| `config.snapSizeLimitInKb` | SnapSizeLimitInKb | `4194304` |
| `config.maxCnxns` | The total number of concurrent connections that can be made to a zookeeper server | `0` |
| `config.maxClientCnxns` | The number of concurrent connections that a single client | `60` |
| `config.minSessionTimeout` | The minimum session timeout in milliseconds that the server will allow the client to negotiate | `4000` |
| `config.maxSessionTimeout` | The maximum session timeout in milliseconds that the server will allow the client to negotiate | `40000` |
| `config.autoPurgeSnapRetainCount` | The number of snapshots to be retained | `3`
| `config.autoPurgePurgeInterval` | The time interval in hours for which the purge task has to be triggered | `1`
| `config.quorumListenOnAllIPs` | Whether Zookeeper server will listen for connections from its peers on all available IP addresses | `false` |
| `config.additionalConfig` | Additional zookeeper coniguration parameters that should be defined in generated zoo.cfg file | `{}` |
| `storageType` | Type of storage that can be used it can take either ephemeral or persistence as value | `persistence` |
| `persistence.reclaimPolicy` | Reclaim policy for persistent volumes | `Delete` |
| `persistence.annotations` | Specifies the annotations to attach to pvcs | `{}` |`
| `persistence.storageClassName` | Storage class for persistent volumes | `` |
| `persistence.volumeSize` | Size of the volume requested for persistent volumes | `20Gi` |
| `ephemeral.emptydirvolumesource.medium` |  What type of storage medium should back the directory. | `""` |
| `ephemeral.emptydirvolumesource.sizeLimit` | Total amount of local storage required for the EmptyDir volume. | `20Gi` |
| `containers` | Application containers run with the zookeeper pod | `[]` |
| `initContainers` | Init Containers to add to the zookeeper pods | `[]` |
| `volumes` | Named volumes that may be accessed by any container in the pod | `[]` |
| `volumeMounts` | Customized volumeMounts for zookeeper container that can be configured to mount volumes to zookeeper container | `[]` |
