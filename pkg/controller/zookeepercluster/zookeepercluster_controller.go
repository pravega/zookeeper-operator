/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package zookeepercluster

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/pravega/zookeeper-operator/pkg/utils"
	"github.com/pravega/zookeeper-operator/pkg/yamlexporter"

	"k8s.io/apimachinery/pkg/labels"

	"github.com/go-logr/logr"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	zookeeperv1beta1 "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 30 * time.Second

var log = logf.Log.WithName("controller_zookeepercluster")

// AddZookeeperReconciler creates a new ZookeeperCluster Controller and adds it
// to the Manager. The Manager will set fields on the Controller and Start it
// when the Manager is Started.
func AddZookeeperReconciler(mgr manager.Manager) error {
	return add(mgr, newZookeeperClusterReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newZookeeperClusterReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZookeeperCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zookeepercluster-controller", mgr,
		controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ZookeeperCluster
	err = c.Watch(&source.Kind{Type: &zookeeperv1beta1.ZookeeperCluster{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper stateful-set secondary resources
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper service secondary resources
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper pod secondary resources
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileZookeeperCluster{}

// ReconcileZookeeperCluster reconciles a ZookeeperCluster object
type ReconcileZookeeperCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client           client.Client
	scheme           *runtime.Scheme
	log              logr.Logger
	skipSTSReconcile int
}

type reconcileFun func(cluster *zookeeperv1beta1.ZookeeperCluster) error

// Reconcile reads that state of the cluster for a ZookeeperCluster object and
// makes changes based on the state read and what is in the ZookeeperCluster.Spec
func (r *ReconcileZookeeperCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues(
		"Request.Namespace", request.Namespace,
		"Request.Name", request.Name)
	r.log.Info("Reconciling ZookeeperCluster")

	// Fetch the ZookeeperCluster instance
	instance := &zookeeperv1beta1.ZookeeperCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile
			// request. Owned objects are automatically garbage collected. For
			// additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	changed := instance.WithDefaults()
	if changed {
		r.log.Info("Setting default settings for zookeeper-cluster")
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}
	for _, fun := range []reconcileFun{
		r.reconcileFinalizers,
		r.reconcileConfigMap,
		r.reconcileStatefulSet,
		r.reconcileClientService,
		r.reconcileHeadlessService,
		r.reconcilePodDisruptionBudget,
		r.reconcileClusterStatus,
	} {
		if err = fun(instance); err != nil {
			return reconcile.Result{}, err
		}
	}
	// Recreate any missing resources every 'ReconcileTime'
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

func (r *ReconcileZookeeperCluster) reconcileStatefulSet(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	sts := zk.MakeStatefulSet(instance)
	if err = controllerutil.SetControllerReference(instance, sts, r.scheme); err != nil {
		return err
	}
	foundSts := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      sts.Name,
		Namespace: sts.Namespace,
	}, foundSts)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Zookeeper StatefulSet",
			"StatefulSet.Namespace", sts.Namespace,
			"StatefulSet.Name", sts.Name)
		err = r.client.Create(context.TODO(), sts)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		foundSTSSize := *foundSts.Spec.Replicas
		newSTSSize := *sts.Spec.Replicas
		if newSTSSize < foundSTSSize {
			// We're dealing with STS scale down
			if !r.isConfigMapInSync(instance) {
				r.log.Info("Skipping StatefulSet reconcile as ConfigMap not updated yet.")
				return nil
			}
			/*
				After updating ConfigMap we need to wait for changes to sync to the volume,
				failing which `zookeeperTeardown.sh` won't get invoked for the pods that are being scaled down
				and these will stay in the ensemble config forever.
				For details see:
				https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/#mounted-configmaps-are-updated-automatically
			*/
			r.skipSTSReconcile++
			if r.skipSTSReconcile < 6 {
				r.log.Info("Waiting for Config Map update to sync...Skipping STS Reconcile")
				return nil
			}
		}
		return r.updateStatefulSet(instance, foundSts, sts)
	}
	return nil
}

func (r *ReconcileZookeeperCluster) updateStatefulSet(instance *zookeeperv1beta1.ZookeeperCluster, foundSts *appsv1.StatefulSet, sts *appsv1.StatefulSet) (err error) {
	r.log.Info("Updating StatefulSet",
		"StatefulSet.Namespace", foundSts.Namespace,
		"StatefulSet.Name", foundSts.Name)
	zk.SyncStatefulSet(foundSts, sts)
	err = r.client.Update(context.TODO(), foundSts)
	if err != nil {
		return err
	}
	instance.Status.Replicas = foundSts.Status.Replicas
	instance.Status.ReadyReplicas = foundSts.Status.ReadyReplicas
	r.skipSTSReconcile = 0
	return nil
}

func (r *ReconcileZookeeperCluster) isConfigMapInSync(instance *zookeeperv1beta1.ZookeeperCluster) bool {
	cm := zk.MakeConfigMap(instance)
	foundCm := &corev1.ConfigMap{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cm.Name,
		Namespace: cm.Namespace,
	}, foundCm)
	if err != nil {
		r.log.Error(err, "Error getting config map.")
		return false
	} else {
		// found config map, now check number of replicas in configMap
		envStr := foundCm.Data["env.sh"]
		splitSlice := strings.Split(envStr, "CLUSTER_SIZE=")
		if len(splitSlice) < 2 {
			r.log.Error(err, "Error: Could not find cluster size in configmap.")
			return false
		}
		cs := strings.TrimSpace(splitSlice[1])
		clusterSize, _ := strconv.Atoi(cs)
		return (int32(clusterSize) == instance.Spec.Replicas)
	}
	return false
}

func (r *ReconcileZookeeperCluster) reconcileClientService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeClientService(instance)
	if err = controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return err
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating new client service",
			"Service.Namespace", svc.Namespace,
			"Service.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		r.log.Info("Updating existing client service",
			"Service.Namespace", foundSvc.Namespace,
			"Service.Name", foundSvc.Name)
		zk.SyncService(foundSvc, svc)
		err = r.client.Update(context.TODO(), foundSvc)
		if err != nil {
			return err
		}
		port := instance.ZookeeperPorts().Client
		instance.Status.InternalClientEndpoint = fmt.Sprintf("%s:%d",
			foundSvc.Spec.ClusterIP, port)
		if foundSvc.Spec.Type == "LoadBalancer" {
			for _, i := range foundSvc.Status.LoadBalancer.Ingress {
				if i.IP != "" {
					instance.Status.ExternalClientEndpoint = fmt.Sprintf("%s:%d",
						i.IP, port)
				}
			}
		} else {
			instance.Status.ExternalClientEndpoint = "N/A"
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) reconcileHeadlessService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeHeadlessService(instance)
	if err = controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return err
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating new headless service",
			"Service.Namespace", svc.Namespace,
			"Service.Name", svc.Name)
		err = r.client.Create(context.TODO(), svc)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		r.log.Info("Updating existing headless service",
			"Service.Namespace", foundSvc.Namespace,
			"Service.Name", foundSvc.Name)
		zk.SyncService(foundSvc, svc)
		err = r.client.Update(context.TODO(), foundSvc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) reconcilePodDisruptionBudget(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	pdb := zk.MakePodDisruptionBudget(instance)
	if err = controllerutil.SetControllerReference(instance, pdb, r.scheme); err != nil {
		return err
	}
	foundPdb := &policyv1beta1.PodDisruptionBudget{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      pdb.Name,
		Namespace: pdb.Namespace,
	}, foundPdb)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating new pod-disruption-budget",
			"PodDisruptionBudget.Namespace", pdb.Namespace,
			"PodDisruptionBudget.Name", pdb.Name)
		err = r.client.Create(context.TODO(), pdb)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	return nil
}

func (r *ReconcileZookeeperCluster) reconcileConfigMap(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	cm := zk.MakeConfigMap(instance)
	if err = controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return err
	}
	foundCm := &corev1.ConfigMap{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      cm.Name,
		Namespace: cm.Namespace,
	}, foundCm)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating a new Zookeeper Config Map",
			"ConfigMap.Namespace", cm.Namespace,
			"ConfigMap.Name", cm.Name)
		err = r.client.Create(context.TODO(), cm)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		r.log.Info("Updating existing config-map",
			"ConfigMap.Namespace", foundCm.Namespace,
			"ConfigMap.Name", foundCm.Name)
		zk.SyncConfigMap(foundCm, cm)
		err = r.client.Update(context.TODO(), foundCm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) reconcileClusterStatus(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	foundPods := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": instance.GetName()})
	listOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.client.List(context.TODO(), listOps, foundPods)
	if err != nil {
		return err
	}
	var (
		readyMembers   []string
		unreadyMembers []string
	)
	for _, p := range foundPods.Items {
		ready := true
		for _, c := range p.Status.ContainerStatuses {
			if !c.Ready {
				ready = false
			}
		}
		if ready {
			readyMembers = append(readyMembers, p.Name)
		} else {
			unreadyMembers = append(unreadyMembers, p.Name)
		}
	}
	instance.Status.Members.Ready = readyMembers
	instance.Status.Members.Unready = unreadyMembers
	r.log.Info("Updating zookeeper status",
		"StatefulSet.Namespace", instance.Namespace,
		"StatefulSet.Name", instance.Name)
	return r.client.Status().Update(context.TODO(), instance)
}

// YAMLExporterReconciler returns a fake Reconciler which is being used for generating YAML files
func YAMLExporterReconciler(zookeepercluster *zookeeperv1beta1.ZookeeperCluster) *ReconcileZookeeperCluster {
	var scheme = scheme.Scheme
	scheme.AddKnownTypes(zookeeperv1beta1.SchemeGroupVersion, zookeepercluster)
	return &ReconcileZookeeperCluster{
		client: fake.NewFakeClient(zookeepercluster),
		scheme: scheme,
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
	sts := zk.MakeStatefulSet(instance)
	if err = controllerutil.SetControllerReference(instance, sts, r.scheme); err != nil {
		return err
	}
	subdir, err := yamlexporter.CreateOutputSubDir(sts.OwnerReferences[0].Kind, sts.Labels["component"])
	return yamlexporter.GenerateOutputYAMLFile(subdir, sts.Kind, sts)
}

// yamlClientService will generates YAML file for zookeeper client service
func (r *ReconcileZookeeperCluster) yamlClientService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeClientService(instance)
	if err = controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return err
	}
	subdir, err := yamlexporter.CreateOutputSubDir(svc.OwnerReferences[0].Kind, "client")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, svc.Kind, svc)
}

// yamlHeadlessService will generates YAML file for zookeeper headless service
func (r *ReconcileZookeeperCluster) yamlHeadlessService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeHeadlessService(instance)
	if err = controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return err
	}
	subdir, err := yamlexporter.CreateOutputSubDir(svc.OwnerReferences[0].Kind, "headless")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, svc.Kind, svc)
}

// yamlPodDisruptionBudget will generates YAML file for zookeeper PDB
func (r *ReconcileZookeeperCluster) yamlPodDisruptionBudget(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	pdb := zk.MakePodDisruptionBudget(instance)
	if err = controllerutil.SetControllerReference(instance, pdb, r.scheme); err != nil {
		return err
	}
	subdir, err := yamlexporter.CreateOutputSubDir(pdb.OwnerReferences[0].Kind, "pdb")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, pdb.Kind, pdb)
}

// yamlConfigMap will generates YAML file for Zookeeper configmap
func (r *ReconcileZookeeperCluster) yamlConfigMap(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	cm := zk.MakeConfigMap(instance)
	if err = controllerutil.SetControllerReference(instance, cm, r.scheme); err != nil {
		return err
	}
	subdir, err := yamlexporter.CreateOutputSubDir(cm.OwnerReferences[0].Kind, "config")
	if err != nil {
		return err
	}
	return yamlexporter.GenerateOutputYAMLFile(subdir, cm.Kind, cm)
}

func (r *ReconcileZookeeperCluster) reconcileFinalizers(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	if instance.Spec.Persistence.VolumeReclaimPolicy != zookeeperv1beta1.VolumeReclaimPolicyDelete {
		return nil
	}
	if instance.DeletionTimestamp.IsZero() {
		if !utils.ContainsString(instance.ObjectMeta.Finalizers, utils.ZkFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, utils.ZkFinalizer)
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return err
			}
		}
		return r.cleanupOrphanPVCs(instance)
	} else {
		if utils.ContainsString(instance.ObjectMeta.Finalizers, utils.ZkFinalizer) {
			if err = r.cleanUpAllPVCs(instance); err != nil {
				return err
			}
			instance.ObjectMeta.Finalizers = utils.RemoveString(instance.ObjectMeta.Finalizers, utils.ZkFinalizer)
			if err = r.client.Update(context.TODO(), instance); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) getPVCCount(instance *zookeeperv1beta1.ZookeeperCluster) (pvcCount int, err error) {
	pvcList, err := r.getPVCList(instance)
	if err != nil {
		return -1, err
	}
	pvcCount = len(pvcList.Items)
	return pvcCount, nil
}

func (r *ReconcileZookeeperCluster) cleanupOrphanPVCs(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	// this check should make sure we do not delete the PVCs before the STS has scaled down
	if instance.Status.ReadyReplicas == instance.Spec.Replicas {
		pvcCount, err := r.getPVCCount(instance)
		if err != nil {
			return err
		}
		r.log.Info("cleanupOrphanPVCs", "PVC Count", pvcCount, "ReadyReplicas Count", instance.Status.ReadyReplicas)
		if pvcCount > int(instance.Spec.Replicas) {
			pvcList, err := r.getPVCList(instance)
			if err != nil {
				return err
			}
			for _, pvcItem := range pvcList.Items {
				// delete only Orphan PVCs
				if utils.IsPVCOrphan(pvcItem.Name, instance.Spec.Replicas) {
					r.deletePVC(pvcItem)
				}
			}
		}
	}
	return nil
}

func (r *ReconcileZookeeperCluster) getPVCList(instance *zookeeperv1beta1.ZookeeperCluster) (pvList corev1.PersistentVolumeClaimList, err error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"app": instance.GetName()},
	})
	pvclistOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: selector,
	}
	pvcList := &corev1.PersistentVolumeClaimList{}
	err = r.client.List(context.TODO(), pvclistOps, pvcList)
	return *pvcList, err
}

func (r *ReconcileZookeeperCluster) cleanUpAllPVCs(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	pvcList, err := r.getPVCList(instance)
	if err != nil {
		return err
	}
	for _, pvcItem := range pvcList.Items {
		r.deletePVC(pvcItem)
	}
	return nil
}

func (r *ReconcileZookeeperCluster) deletePVC(pvcItem corev1.PersistentVolumeClaim) {
	pvcDelete := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcItem.Name,
			Namespace: pvcItem.Namespace,
		},
	}
	r.log.Info("Deleting PVC", "With Name", pvcItem.Name)
	err := r.client.Delete(context.TODO(), pvcDelete)
	if err != nil {
		r.log.Error(err, "Error deleteing PVC.", "Name", pvcDelete.Name)
	}
}
