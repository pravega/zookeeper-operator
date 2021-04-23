package zookeepercluster

import (
	"context"
	"fmt"

	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	zookeeperv1beta1 "github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/utils"
)

func (r *ReconcileZookeeperCluster) reconcileClusterStatus(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	if instance.Status.IsClusterInUpgradingState() || instance.Status.IsClusterInUpgradeFailedState() {
		return nil
	}
	instance.Status.Init()
	// foundPods := &corev1.PodList{}
	foundPods := &statefulpodv1.StatefulPodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": instance.GetName()})
	listOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), foundPods, listOps)
	if err != nil {
		return err
	}
	var (
		readyMembers   []string
		unreadyMembers []string
	)
	for _, item := range foundPods.Items {
		for _, pod := range item.Status.PodStatusMes {
			ready := true
			if pod.Status != corev1.PodRunning {
				ready = false
			}
			if ready {
				readyMembers = append(readyMembers, pod.PodName)
			} else {
				unreadyMembers = append(unreadyMembers, pod.PodName)
			}
		}
	}

	instance.Status.Members.Ready = readyMembers
	instance.Status.Members.Unready = unreadyMembers
	instance.Status.ReadyReplicas = int32(len(readyMembers))

	//If Cluster is in a ready state...
	if instance.Spec.Replicas == instance.Status.ReadyReplicas && (!instance.Status.MetaRootCreated) {
		r.log.Info("Cluster is Ready, Creating ZK Metadata...")
		zkUri := utils.GetZkServiceUri(instance)
		err := r.zkClient.Connect(zkUri)
		if err != nil {
			return fmt.Errorf("Error creating cluster metaroot. Connect to zk failed %v", err)
		}
		defer r.zkClient.Close()
		metaPath := utils.GetMetaPath(instance)
		r.log.Info("Connected to zookeeper:", "ZKUri", zkUri, "Creating Path", metaPath)
		if err := r.zkClient.CreateNode(instance, metaPath); err != nil {
			return fmt.Errorf("Error creating cluster metadata path %s, %v", metaPath, err)
		}
		r.log.Info("Metadata znode created.")
		instance.Status.MetaRootCreated = true
	}
	r.log.Info("Updating zookeeper status",
		"StatefulPod.Namespace", instance.Namespace,
		"StatefulPod.Name", instance.Name)
	if instance.Status.ReadyReplicas == instance.Spec.Replicas {
		instance.Status.SetPodsReadyConditionTrue()
	} else {
		instance.Status.SetPodsReadyConditionFalse()
	}
	if instance.Status.CurrentVersion == "" && instance.Status.IsClusterInReadyState() {
		instance.Status.CurrentVersion = instance.Spec.Image.Tag
	}
	return r.client.Status().Update(context.TODO(), instance)
}
