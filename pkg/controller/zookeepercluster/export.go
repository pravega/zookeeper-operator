package zookeepercluster

import (
	"fmt"

	zookeeperv1beta1 "github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/yamlexporter"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/zk"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// YAMLExporterReconciler returns a fake Reconciler which is being used for generating YAML files
func YAMLExporterReconciler(zookeepercluster *zookeeperv1beta1.ZookeeperCluster) *ReconcileZookeeperCluster {
	var scheme = scheme.Scheme
	scheme.AddKnownTypes(zookeeperv1beta1.SchemeGroupVersion, zookeepercluster)
	return &ReconcileZookeeperCluster{
		client:   fake.NewFakeClient(zookeepercluster),
		scheme:   scheme,
		zkClient: new(zk.DefaultZookeeperClient),
	}
}

// GenerateYAML generated secondary resource of ZookeeperCluster resources YAML files
func (r *ReconcileZookeeperCluster) GenerateYAML(inst *zookeeperv1beta1.ZookeeperCluster) error {
	if inst.WithDefaults() {
		fmt.Println("set default values")
	}
	for _, fun := range []reconcileFun{
		r.yamlConfigMap,
		r.yamlStatefulSet,
		r.yamlClientService,
		r.yamlHeadlessService,
		r.yamlPodDisruptionBudget,
	} {
		if err := fun(inst); err != nil {
			return err
		}
	}
	return nil
}

// yamlStatefulSet will generates YAML file for StatefulSet
func (r *ReconcileZookeeperCluster) yamlStatefulSet(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	statefulPod := zk.MakeStatefulPod(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", statefulPod.Labels["component"])
	return yamlexporter.GenerateOutputYAMLFile(subdir, statefulPod.Kind, statefulPod)
}

// yamlClientService will generates YAML file for zookeeper client service
func (r *ReconcileZookeeperCluster) yamlClientService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeClientService(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", "client")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, svc.Kind, svc)
}

// yamlHeadlessService will generates YAML file for zookeeper headless service
func (r *ReconcileZookeeperCluster) yamlHeadlessService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeHeadlessService(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", "headless")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, svc.Kind, svc)
}

// yamlPodDisruptionBudget will generates YAML file for zookeeper PDB
func (r *ReconcileZookeeperCluster) yamlPodDisruptionBudget(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	pdb := zk.MakePodDisruptionBudget(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", "pdb")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, pdb.Kind, pdb)
}

// yamlConfigMap will generates YAML file for Zookeeper configmap
func (r *ReconcileZookeeperCluster) yamlConfigMap(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	cm := zk.MakeConfigMap(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", "config")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, cm.Kind, cm)
}
