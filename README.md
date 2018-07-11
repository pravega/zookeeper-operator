# Zookeeper Operator
>**This operator is in WIP state and subject to (breaking) changes.**

This Operator runs a Zookeeper 3.5 cluster, and uses Zookeeper dynamic reconfiguration to handle node membership.

The operator itself is built with the: https://github.com/operator-framework/operator-sdk

## Build Requirements:
 - Install the Operator SDK first: https://github.com/operator-framework/operator-sdk#quick-start

## Usage:

```bash
mkdir -p $GOPATH/src/github.com/pravega
cd $GOPATH/src/github.com/pravega
git clone git@github.com:pravega/zookeeper-operator.git
cd zookeeper-operator
```

### Get the operator Docker image

#### a. Build the image yourself

```bash
operator-sdk build pravega/zookeeper-operator
docker tag pravega/zookeeper-operator ${your-operator-image-tag}:latest
docker push ${your-operator-image-tag}:latest
```

#### b. Use the image from Docker Hub

```bash
# No addition steps needed
```

### Install the Kubernetes resources

```bash
# Create Operator deployment, Roles, Service Account, and Custom Resource Definition for
#   a Zookeeper cluster.
$ kubectl apply -f deploy

# View the zookeeper-operator Pod
$ kubectl get pod
NAME                                  READY     STATUS              RESTARTS   AGE
zookeeper-operator-5c7b8cfd85-ttb5g   1/1       Running             0          5m
```

### The Zookeeper Custom Resource

With this YAML template you can install a 3 node Zookeeper Cluster easily into your Kubernetes cluster:

```yaml
apiVersion: "zookeeper.pravega.io/v1beta1"
kind: "ZookeeperCluster"
metadata:
  name: "example"
spec:
  size: 3
```

After creating, you can view the cluster:

```bash
# View the new zookeeper cluster instance
$ kubectl get zk
NAME      AGE
example   2s

# View what it's made of
$ kubectl get all -l app=example
NAME            READY     STATUS              RESTARTS   AGE
pod/example-0   1/1       Running             0          51m
pod/example-1   1/1       Running             0          55m
pod/example-2   1/1       Running             0          58m

NAME                       TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
service/example-client     ClusterIP   x.x.x.x         <none>        2181/TCP            51m
service/example-headless   ClusterIP   None            <none>        2888/TCP,3888/TCP   51m

NAME                       DESIRED   CURRENT   AGE
statefulset.apps/example   3         3         58m

# There are a few other things here, like a configmap, poddisruptionbudget, etc...
```
