## Setting up RBAC for zookeeper stateful set
### Use non-default service accounts

From zookeeper operator version `0.2.9` onwards, support is added to configure non deault service accounts for zookeeper pods.

Below are the steps

1. Create a service account in the namespace where you want to deploy zookeeper cluster

```
apiVersion: v1
kind: ServiceAccount
metadata:
  name: zookeeper
```

2. Create `Role` and `ClusterRole` with the minimum required permissions. Make sure to update the `default` namespace if you are deploying zookeeper to a custom namespace.

```
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: zookeeper
  namespace: "default"
rules:
- apiGroups: ["zookeeper.pravega.io"]
  resources: ["*"]
  verbs: ["get"]
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get"]
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: zookeeper
rules:
- apiGroups: [""]
  resources: ["nodes"]
  verbs: ["get"]
```
3. Bind roles and cluster roles to the service account.

```
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: zookeeper
subjects:
- kind: ServiceAccount
  name: zookeeper
roleRef:
  kind: Role
  name: zookeeper
  apiGroup: rbac.authorization.k8s.io
---
kind: ClusterRoleBinding
apiVersion: rbac.authorization.k8s.io/v1beta1
metadata:
  name: zookeeper
subjects:
- kind: ServiceAccount
  name: zookeeper
  namespace: default
roleRef:
  kind: ClusterRole
  name: zookeeper
  apiGroup: rbac.authorization.k8s.io
```

4. Create zookeeper cluster
  ```
  apiVersion: zookeeper.pravega.io/v1beta1
  kind: ZookeeperCluster
  metadata:
    name: zookeeper
  spec:
    replicas: 3
    image:
      repository: pravega/zookeeper
      tag: 0.2.9
    pod:
      serviceAccountName: zookeeper  
    storageType: persistence
    persistence:
      reclaimPolicy: Delete
      spec:
        storageClassName: "standard"
        resources:
          requests:
            storage: 20Gi
```
