package zk

import "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"

func Reconcile(zk *v1beta1.ZookeeperCluster) (err error) {
	zk = zk.DeepCopy()
	zk.WithDefaults()

	deploy(zk)
	syncClusterSize(zk)

	return nil
}
