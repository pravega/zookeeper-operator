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
    * [Upgrade a Zookeeper Cluster](#upgrade-a-zookeeper-cluster)
    * [Uninstall the Zookeeper Cluster](#uninstall-the-zookeeper-cluster)
    * [Uninstall the Operator](#uninstall-the-operator)
 * [Development](#development)
    * [Build the Operator Image](#build-the-operator-image)
    * [Direct Access to Cluster](#direct-access-to-the-cluster)
    * [Run the Operator Locally](#run-the-operator-locally)
    * [Installation on GKE](#installation-on-google-kubernetes-engine)



### Overview

This operator runs a Zookeeper 3.5 cluster, and uses Zookeeper dynamic reconfiguration to handle node membership.

The operator itself is built with the [Operator framework](https://github.com/operator-framework/operator-sdk).

## Requirements

- Access to a Kubernetes v1.9.0+ cluster

## Usage

### Install the operator

> Note: if you are running on Google Kubernetes Engine (GKE), please [check this first](#installation-on-google-kubernetes-engine).

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

Create a Yaml file called `zk.yaml` with the following content to install a 3-node Zookeeper cluster.

```yaml
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "example"
spec:
  size: 3
```

```
$ kubectl create -f zk.yaml
```

Verify that the cluster instances and its components are running.

```
$ kubectl get zk
NAME      AGE
example   15s
```

```
$ kubectl get all -l app=example
NAME                   DESIRED   CURRENT   AGE
statefulsets/example   3         3         2m

NAME           READY     STATUS    RESTARTS   AGE
po/example-0   1/1       Running   0          2m
po/example-1   1/1       Running   0          1m
po/example-2   1/1       Running   0          1m

NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
svc/example-client     ClusterIP   10.31.243.173   <none>        2181/TCP            2m
svc/example-headless   ClusterIP   None            <none>        2888/TCP,3888/TCP   2m
```

### Upgrade a Zookeeper cluster

To initiate an upgrade process, a user has to update the `spec.image.tag` field of the `ZookeeperCluster` custom resource. This can be done in three different ways using the `kubectl` command.
1. `kubectl edit zk <name>`, modify the `tag` value in the YAML resource, save, and exit.
2. If you have the custom resource defined in a local YAML file, e.g. `zk.yaml`, you can modify the `tag` value, and reapply the resource with `kubectl apply -f zk.yaml`.
3. `kubectl patch zk <name> --type='json' -p='[{"op": "replace", "path": "/spec/image/tag", "value": "X.Y.Z"}]'`.

After the `tag` field is updated, the StatefulSet will detect the version change and it will trigger the upgrade process.

To detect whether a `ZookeeperCluster` upgrade is in progress or not, check the output of the command `kubectl get sts <name> -o yaml`. The output of this command contains the following entries

```
status:
  collisionCount: 0
  currentReplicas: 2
  currentRevision: example-75d4b645c8
  observedGeneration: 4
  readyReplicas: 3
  replicas: 3
  updateRevision: example-5d9dfdb787
```

If the values of the fields `currentRevision` and `updateRevision` are different, it indicates that the `ZookeeperCluster` is currently undergoing an upgrade. The value of `currentRevision` is set to the value of `updateRevision` when the upgrade is complete.

```
Note: The value of the tag field should not be modified while an upgrade is already in progress.
```

### Uninstall the Zookeeper cluster

```
$ kubectl delete -f zk.yaml
```

### Uninstall the operator

> Note that the Zookeeper clusters managed by the Zookeeper operator will NOT be deleted even if the operator is uninstalled.

To delete all clusters, delete all cluster CR objects before uninstalling the operator.

```
$ kubectl delete -f deploy/default_ns
// or, depending on how you deployed it
$ kubectl delete -f deploy/all_ns
```

## Development

### Build the operator image

Requirements:
  - Go 1.12+

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

For debugging and development you might want to access the Zookeeper cluster directly. For example, if you created the cluster with name `example` in the `default` namespace you can forward the Zookeeper port from any of the pods (e.g. `example-0`) as follows:

```
$ kubectl port-forward -n default example-0 2181:2181
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


#### Zookeeper YAML  Exporter

Zookeeper Exporter is a binary which is used to generate YAML file for all the secondary resources which Zookeeper Operator deploys to the Kubernetes Cluster. It takes ZookeeperCluster resource YAML file as input and generates bunch of secondary resources YAML files. The generated output look like the following:

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


#How to build Zookeeper Operator
When you build Operator, the Exporter is built along with it.
make build-go - will build both Operator as well as Exporter.

#How to use exporter
Just run zookeeper-exporter binary with -help option. It will guide you to input ZookeeperCluster YAML file. There are couple of more options to specify.
Example: ./zookeeper-exporter -i ./ZookeeperCluster.yaml -o .
