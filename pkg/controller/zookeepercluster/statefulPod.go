package zookeepercluster

import (
	"context"
	"fmt"
	"strconv"

	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	zookeeperv1beta1 "github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/utils"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/zk"
)

func (r *ReconcileZookeeperCluster) reconcileStatefulPod(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	// we cannot upgrade if cluster is in UpgradeFailed
	if instance.Status.IsClusterInUpgradeFailedState() {
		return nil
	}
	statusFulPod := zk.MakeStatefulPod(instance)
	if err = controllerutil.SetControllerReference(instance, statusFulPod, r.scheme); err != nil {
		return err
	}
	foundStatefulPod := &statefulpodv1.StatefulPod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      statusFulPod.Name,
		Namespace: statusFulPod.Namespace,
	}, foundStatefulPod)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Zookeeper StatefulPod",
			"StatefulPod.Namespace", statusFulPod.Namespace,
			"StatefulPod.Name", statusFulPod.Name)
		return r.client.Create(context.TODO(), statusFulPod)
	} else if err != nil {
		return err
	} else {
		foundSTSSize := *foundStatefulPod.Spec.Size
		newSTSSize := *statusFulPod.Spec.Size
		if newSTSSize != foundSTSSize {
			zkUri := utils.GetZkServiceUri(instance)
			err = r.zkClient.Connect(zkUri)
			if err != nil {
				return fmt.Errorf("Error storing cluster size %v", err)
			}
			defer r.zkClient.Close()
			r.log.Info("Connected to ZK", "ZKURI", zkUri)

			path := utils.GetMetaPath(instance)
			version, err := r.zkClient.NodeExists(path)
			if err != nil {
				return fmt.Errorf("Error doing exists check for znode %s: %v", path, err)
			}

			data := "CLUSTER_SIZE=" + strconv.Itoa(int(newSTSSize))
			r.log.Info("Updating Cluster Size.", "New Data:", data, "Version", version)
			err = r.zkClient.UpdateNode(path, data, version)
			if err != nil {
				return err
			}
			return r.updateStatefulPod(instance, foundStatefulPod, statusFulPod)
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) updateStatefulPod(instance *zookeeperv1beta1.ZookeeperCluster,
	foundSts *statefulpodv1.StatefulPod, statusFulPod *statefulpodv1.StatefulPod) (err error) {
	r.log.Info("Updating StatefulPod",
		"StatefulPod.Namespace", foundSts.Namespace,
		"StatefulPod.Name", foundSts.Name)
	zk.SyncStatefulPod(foundSts, statusFulPod)

	err = r.client.Update(context.TODO(), foundSts)
	if err != nil {
		return err
	}
	instance.Status.Replicas = *foundSts.Spec.Size
	return nil
}
