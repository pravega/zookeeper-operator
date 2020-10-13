# Upgrading Operator
zookeeperoperator can be upgraded to a version **[VERSION]** via helm using the following command

```
$ helm upgrade [ZOOKEEPER_OPERATOR_RELEASE_NAME] pravega/zookeeper-operator --version=[VERSION]
```
The zookeeper operator with deployment name **[DEPLOYMENT_NAME]** can also be upgraded manually by modifying the image tag using kubectl edit, patch or apply
```
$ kubectl edit deploy [DEPLOYMENT_NAME]
```
> Note: If you are upgrading zookeeper operator version to 0.2.9 or above manually, clusterrole has to be updated to include serviceaccounts. After updating clusterroles, zookeeper operator pod has to be restarted for the changes to take effect.

```
- apiGroups:
  - ""
  resources:
  - pods
  - services
  - endpoints
  - persistentvolumeclaims
  - events
  - configmaps
  - secrets
  - serviceaccounts
  verbs:
  - "*"
```
