# Zookeeper Operator
[![Build Status](https://travis-ci.org/pravega/zookeeper-operator.svg?branch=master)](https://travis-ci.org/pravega/zookeeper-operator)
### Project status: alpha

The project is currently alpha. While no breaking API changes are currently planned, we reserve the right to address bugs and change the API before the project is declared stable.

## Table of Contents

 * [Overview](#overview)
 * [Requirements](#requirements)
 * [Usage](#usage)    
    * [Installation of the Operator](#install-the-operator)
    * [Deploy a sample Zookeeper Cluster](#deploy-a-sample-zookeeper-cluster)
    * [Deploy a sample ZooKeeper Cluster with Ephemeral Storage](#Deploy-a-sample-zookeeper-cluster-with-ephemeral-storage)
    * [Deploy a sample Zookeeper Cluster to a cluster using Istio](#deploy-a-sample-zookeeper-cluster-with-istio)
    * [Upgrade a Zookeeper Cluster](#upgrade-a-zookeeper-cluster)
    * [Uninstall the Zookeeper Cluster](#uninstall-the-zookeeper-cluster)
    * [Upgrade the Zookeeper Operator](#upgrade-the-operator)
    * [Uninstall the Operator](#uninstall-the-operator)
    * [The AdminServer](#the-adminserver)
 * [Development](#development)
    * [Build the Operator Image](#build-the-operator-image)
    * [Direct Access to Cluster](#direct-access-to-the-cluster)
    * [Run the Operator Locally](#run-the-operator-locally)
    * [Installation on GKE](#installation-on-google-kubernetes-engine)
    * [Installation on Minikube](#installation-on-minikube)


### Overview

This operator runs a Zookeeper 3.6.3 cluster, and uses Zookeeper dynamic reconfiguration to handle node membership.

The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

## Requirements

- Access to a Kubernetes v1.15.0+ cluster

## Usage

We recommend using our [helm charts](charts) for all installation and upgrades. Since version 0.2.8 onwards, the helm charts for zookeeper operator and zookeeper cluster are published in [https://charts.pravega.io](https://charts.pravega.io/). To add this repository to your Helm repos, use the following command
```
helm repo add pravega https://charts.pravega.io
```
However there are manual deployment and upgrade options available as well.

### Install the operator

> Note: if you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

#### Install via helm

To understand how to deploy the zookeeper operator using helm, refer to [this](charts/zookeeper-operator#installing-the-chart).

#### Manual deployment

Register the `ZookeeperCluster` custom resource definition (CRD).

```
$ kubectl create -f deploy/crds
```

You can choose to enable Zookeeper operator for all namespaces or just for a specific namespace. The example is using the `default` namespace, but feel free to edit the Yaml files and use a different namespace.

Create the operator role and role binding.

```
// default namespace
$ kubectl create -f deploy/default_ns/rbac.yaml

// all namespaces
$ kubectl create -f deploy/all_ns/rbac.yaml
```

Deploy the Zookeeper operator.

```
// default namespace
$ kubectl create -f deploy/default_ns/operator.yaml

// all namespaces
$ kubectl create -f deploy/all_ns/operator.yaml
```

Verify that the Zookeeper operator is running.

```
$ kubectl get deploy
NAME                 DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
zookeeper-operator   1         1         1            1           12m
```

### Deploy a sample Zookeeper cluster

#### Install via helm

To understand how to deploy a sample zookeeper cluster using helm, refer to [this](charts/zookeeper#installing-the-chart).

#### Manual deployment

Create a Yaml file called `zk.yaml` with the following content to install a 3-node Zookeeper cluster.

```yaml
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "zookeeper"
spec:
  replicas: 3
```

```
$ kubectl create -f zk.yaml
```

After a couple of minutes, all cluster members should become ready.

```
$ kubectl get zk

NAME        REPLICAS   READY REPLICAS    VERSION   DESIRED VERSION   INTERNAL ENDPOINT    EXTERNAL ENDPOINT   AGE
zookeeper   3          3                 0.2.8     0.2.8             10.100.200.18:2181   N/A                 94s
```
>Note: when the Version field is set as well as Ready Replicas are equal to Replicas that signifies our cluster is in Ready state

Additionally, check the output of describe command which should show the following cluster condition

```
$ kubectl describe zk

Conditions:
  Last Transition Time:    2020-05-18T10:17:03Z
  Last Update Time:        2020-05-18T10:17:03Z
  Status:                  True
  Type:                    PodsReady

```
>Note: User should wait for the Pods Ready condition to be True

```
$ kubectl get all -l app=zookeeper
NAME                     DESIRED   CURRENT   AGE
statefulsets/zookeeper   3         3         2m

NAME             READY     STATUS    RESTARTS   AGE
po/zookeeper-0   1/1       Running   0          2m
po/zookeeper-1   1/1       Running   0          1m
po/zookeeper-2   1/1       Running   0          1m

NAME                     TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
svc/zookeeper-client     ClusterIP   10.31.243.173   <none>        2181/TCP            2m
svc/zookeeper-headless   ClusterIP   None            <none>        2888/TCP,3888/TCP   2m
```

>Note: If you want to configure non default service accounts to zookeeper pods, set the service account inside pod.This support is added from zookeeper operator version `0.2.9` onwards.

```
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "example"
spec:
  pod:
    serviceAccountName: "zookeeper"
```

### Deploy a sample Zookeeper cluster with Ephemeral storage

Create a Yaml file called `zk.yaml` with the following content to install a 3-node Zookeeper cluster.

```yaml
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "example"
spec:
  replicas: 3        
  storageType: ephemeral
```

```
$ kubectl create -f zk.yaml
```

After a couple of minutes, all cluster members should become ready.

```
$ kubectl get zk

NAME      REPLICAS   READY REPLICAS   VERSION   DESIRED VERSION   INTERNAL ENDPOINT    EXTERNAL ENDPOINT   AGE
example   3          3                 0.2.7     0.2.7             10.100.200.18:2181   N/A                 94s
```
>Note: User should only provide value for either the field persistence or ephemeral in the spec if none of the values is specified default is persistence

>Note: In case of ephemeral storage, the cluster may not be able to come back up if more than quorum number of nodes are restarted simultaneously.

>Note: In case of ephemeral storage, there will be loss of data when the node gets restarted.

### Deploy a sample Zookeeper cluster with Istio
Create a Yaml file called `zk-with-istio.yaml` with the following content to install a 3-node Zookeeper cluster.

```yaml
apiVersion: zookeeper.pravega.io/v1beta1
kind: ZookeeperCluster
metadata:
  name: zk-with-istio
spec:
  replicas: 3
  config:
    initLimit: 10
    tickTime: 2000
    syncLimit: 5
    quorumListenOnAllIPs: true
```

```
$ kubectl create -f zk-with-istio.yaml
```

### Upgrade a Zookeeper cluster

#### Trigger the upgrade via helm

To understand how to upgrade the zookeeper cluster using helm, refer to [this](charts/zookeeper#upgrading-the-chart).

#### Trigger the upgrade manually

To initiate an upgrade process manually, a user has to update the `spec.image.tag` field of the `ZookeeperCluster` custom resource. This can be done in three different ways using the `kubectl` command.
1. `kubectl edit zk <name>`, modify the `tag` value in the YAML resource, save, and exit.
2. If you have the custom resource defined in a local YAML file, e.g. `zk.yaml`, you can modify the `tag` value, and reapply the resource with `kubectl apply -f zk.yaml`.
3. `kubectl patch zk <name> --type='json' -p='[{"op": "replace", "path": "/spec/image/tag", "value": "X.Y.Z"}]'`.

After the `tag` field is updated, the StatefulSet will detect the version change and it will trigger the upgrade process.

To detect whether a `ZookeeperCluster` upgrade is in progress or not, check the output of the command `kubectl describe zk`. Output of this command should contain the following entries

```
$ kubectl describe zk

status:
Last Transition Time:    2020-05-18T10:25:12Z
Last Update Time:        2020-05-18T10:25:12Z
Message:                 0
Reason:                  Updating Zookeeper
Status:                  True
Type:                    Upgrading
```
Additionally, the Desired Version will be set to the version that we are upgrading our cluster to.

```
$ kubectl get zk

NAME            REPLICAS   READY REPLICAS   VERSION   DESIRED VERSION   INTERNAL ENDPOINT     EXTERNAL ENDPOINT   AGE
zookeeper       3          3                0.2.6     0.2.7             10.100.200.126:2181   N/A                 11m

```
Once the upgrade completes, the Version field is set to the Desired Version, as shown below

```
$ kubectl get zk

NAME            REPLICAS   READY REPLICAS   VERSION   DESIRED VERSION   INTERNAL ENDPOINT     EXTERNAL ENDPOINT   AGE
zookeeper       3          3                0.2.7     0.2.7             10.100.200.126:2181   N/A                 11m


```
Additionally, the Upgrading status is set to False and PodsReady status is set to True, which signifies that the upgrade has completed, as shown below

```
$ kubectl describe zk

Status:
  Conditions:
    Last Transition Time:    2020-05-18T10:28:22Z
    Last Update Time:        2020-05-18T10:28:22Z
    Status:                  True
    Type:                    PodsReady
    Last Transition Time:    2020-05-18T10:28:22Z
    Last Update Time:        2020-05-18T10:28:22Z
    Status:                  False
    Type:                    Upgrading
```
>Note: The value of the tag field should not be modified while an upgrade is already in progress.

### Upgrade the Operator

For upgrading the zookeeper operator check the document [operator-upgrade](doc/operator-upgrade.md)

### Uninstall the Zookeeper cluster

#### Uninstall via helm

Refer to [this](charts/zookeeper#uninstalling-the-chart).

#### Manual uninstall

```
$ kubectl delete -f zk.yaml
```

### Uninstall the operator

> Note that the Zookeeper clusters managed by the Zookeeper operator will NOT be deleted even if the operator is uninstalled.

#### Uninstall via helm

Refer to [this](charts/zookeeper-operator#uninstalling-the-chart).

#### Manual uninstall

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy/default_ns
// or, depending on how you deployed it
$ kubectl delete -f deploy/all_ns
```

### The AdminServer
The AdminServer is an embedded Jetty server that provides an HTTP interface to the four letter word commands. This port is made accessible to the outside world via the AdminServer service.
By default, the server is started on port 8080, but this configuration can be modified by providing the desired port number within the values.yaml file of the zookeeper cluster charts
```
ports:
   - containerPort: 8118
     name: admin-server
```
This would bring up the AdminServer service on port 8118 as shown below
```
$ kubectl get svc
NAME                                TYPE           CLUSTER-IP       EXTERNAL-IP    PORT(S)
zookeeper-admin-server              LoadBalancer   10.100.200.104   10.243.39.62   8118:30477/TCP
```
The commands are issued by going to the URL `/commands/<command name>`, e.g. `http://10.243.39.62:8118/commands/stat`
The list of available commands are
```
/commands/configuration
/commands/connection_stat_reset
/commands/connections
/commands/dirs
/commands/dump
/commands/environment
/commands/get_trace_mask
/commands/hash
/commands/initial_configuration
/commands/is_read_only
/commands/last_snapshot
/commands/leader
/commands/monitor
/commands/observer_connection_stat_reset
/commands/observers
/commands/ruok
/commands/server_stats
/commands/set_trace_mask
/commands/stat_reset
/commands/stats
/commands/system_properties
/commands/voting_view
/commands/watch_summary
/commands/watches
/commands/watches_by_path
/commands/zabstate
```

## Development

### Build the operator image

Requirements:
  - Go 1.16+

Use the `make` command to build the Zookeeper operator image.

```
$ make build
```
That will generate a Docker image with the format
`<latest_release_tag>-<number_of_commits_after_the_release>` (it will append-dirty if there are uncommitted changes). The image will also be tagged as `latest`.

Example image after running `make build`.

The Zookeeper operator image will be available in your Docker environment.

```
$ docker images pravega/zookeeper-operator

REPOSITORY                    TAG              IMAGE ID        CREATED         SIZE   

pravega/zookeeper-operator    0.1.1-3-dirty    2b2d5bcbedf5    10 minutes ago  41.7MB

pravega/zookeeper-operator    latest           2b2d5bcbedf5    10 minutes ago  41.7MB

```
Optionally push it to a Docker registry.

```
docker tag pravega/zookeeper-operator [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/zookeeper-operator
docker push [REGISTRY_HOST]:[REGISTRY_PORT]/pravega/zookeeper-operator
```

where:

- `[REGISTRY_HOST]` is your registry host or IP (e.g. `registry.example.com`)
- `[REGISTRY_PORT]` is your registry port (e.g. `5000`)

### Direct access to the cluster

For debugging and development you might want to access the Zookeeper cluster directly. For example, if you created the cluster with name `zookeeper` in the `default` namespace you can forward the Zookeeper port from any of the pods (e.g. `zookeeper-0`) as follows:

```
$ kubectl port-forward -n default zookeeper-0 2181:2181
```

### Run the operator locally

You can run the operator locally to help with development, testing, and debugging tasks.

The following command will run the operator locally with the default Kubernetes config file present at `$HOME/.kube/config`. Use the `--kubeconfig` flag to provide a different path.

```
$ operator-sdk up local
```

### Installation on Google Kubernetes Engine

The Operator requires elevated privileges in order to watch for the custom resources.

According to Google Container Engine docs:

> Ensure the creation of RoleBinding as it grants all the permissions included in the role that we want to create. Because of the way Container Engine checks permissions when we create a Role or ClusterRole.
>
> An example workaround is to create a RoleBinding that gives your Google identity a cluster-admin role before attempting to create additional Role or ClusterRole permissions.
>
> This is a known issue in the Beta release of Role-Based Access Control in Kubernetes and Container Engine version 1.6.

On GKE, the following command must be run before installing the operator, replacing the user with your own details.

```
$ kubectl create clusterrolebinding your-user-cluster-admin-binding --clusterrole=cluster-admin --user=your.google.cloud.email@example.org
```

### Installation on Minikube

#### Minikube Setup
To setup minikube locally you can follow the steps mentioned [here](https://github.com/pravega/pravega/wiki/Kubernetes-Based-System-Test-Framework#minikube-setup).

Once minikube setup is complete, `minikube start` will create a minikube VM.

#### Cluster Deployment
First install the zookeeper operator in either of the ways mentioned [here](#install-the-operator).
Since minikube provides a single node Kubernetes cluster which has a low resource provisioning, we provide a simple way to install a small zookeeper cluster on a minikube environment using the following command.

```
helm install zookeeper charts/zookeeper --values charts/zookeeper/values/minikube.yaml
```

#### Zookeeper YAML  Exporter

Zookeeper Exporter is a binary which is used to generate YAML file for all the secondary resources which Zookeeper Operator deploys to the Kubernetes Cluster. It takes ZookeeperCluster resource YAML file as input and generates bunch of secondary resources YAML files. The generated output look like the following:

```
>tree  ZookeeperCluster/
ZookeeperCluster/
├── client
│   └── Service.yaml
├── config
│   └── ConfigMap.yaml
├── headless
│   └── Service.yaml
├── pdb
│   └── PodDisruptionBudget.yaml
└── zk
    └── StatefulSet.yaml
```

##### How to build Zookeeper Operator

When you build Operator, the Exporter is built along with it.
`make build-go` - will build both Operator as well as Exporter.

##### How to use exporter

Just run zookeeper-exporter binary with -help option. It will guide you to input ZookeeperCluster YAML file. There are couple of more options to specify.
Example: `./zookeeper-exporter -i ./ZookeeperCluster.yaml -o .`
