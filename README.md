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
    * [Deploy a sample Zookeeper Cluster to a cluster using Istio](#deploy-a-sample-zookeeper-cluster-with-istio)
    * [Upgrade a Zookeeper Cluster](#upgrade-a-zookeeper-cluster)
    * [Uninstall the Zookeeper Cluster](#uninstall-the-zookeeper-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
 * [Development](#development)
    * [Build the Operator Image](#build-the-operator-image)
    * [Direct Access to Cluster](#direct-access-to-the-cluster)
    * [Run the Operator Locally](#run-the-operator-locally)
    * [Installation on GKE](#installation-on-google-kubernetes-engine)
    * [Installation on Minikube](#installation-on-minikube)


### Overview

This operator runs a Zookeeper 3.6.1 cluster, and uses Zookeeper dynamic reconfiguration to handle node membership.

The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

## Requirements

- Access to a Kubernetes v1.15.0+ cluster

## Usage

We recommend using our [helm charts](charts) for all installation and upgrades. However there are manual deployment and upgrade options available as well.

### Install the operator

> Note: if you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

#### Install via helm

Use helm to quickly deploy a zookeeper operator with the release name `zookeeper-operator`.

```
$ helm install zookeeper-operator charts/zookeeper-operator
```

#### Manual deployment

Register the `ZookeeperCluster` custom resource definition (CRD).

```
$ kubectl create -f deploy/crds/zookeeper_v1beta1_zookeepercluster_crd.yaml
```

You can choose to enable Zookeeper operator for all namespaces or just for the a specific namespace. The example is using the `default` namespace, but feel free to edit the Yaml files and use a different namespace.

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

Helm can be used to install a sample zookeeper cluster with the release name `zookeeper`.

```
$ helm install zookeeper charts/zookeeper
```

Check out the [zookeeper helm charts](charts/zookeeper) for the complete list of configurable parameters.

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

The upgrade can be triggered via helm using the following command
```
$ helm upgrade zookeeper <location of modified charts> --timeout 600s
```
Here `zookeeper` is the release name of the zookeeper cluster.

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

### Uninstall the Zookeeper cluster

#### Uninstall via helm

```
$ helm uninstall zookeeper
```
Here `zookeeper` is the zookeeper cluster release name.

#### Manual uninstall

```
$ kubectl delete -f zk.yaml
```

### Uninstall the operator

> Note that the Zookeeper clusters managed by the Zookeeper operator will NOT be deleted even if the operator is uninstalled.

#### Uninstall via helm

```
$ helm uninstall zookeeper-operator
```
Here `zookeeper-operator` is the operator release name.

#### Manual uninstall

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy/default_ns
// or, depending on how you deployed it
$ kubectl delete -f deploy/all_ns
```

## Development

### Build the operator image

Requirements:
  - Go 1.13+

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
