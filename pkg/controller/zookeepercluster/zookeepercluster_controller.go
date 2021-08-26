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
	"time"

	"github.com/operator-framework/operator-sdk/pkg/predicate"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/pravega/zookeeper-operator/pkg/utils"
	"github.com/pravega/zookeeper-operator/pkg/yamlexporter"

	"github.com/go-logr/logr"
	zookeeperv1beta1 "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	return &ReconcileZookeeperCluster{client: mgr.GetClient(), scheme: mgr.GetScheme(), zkClient: new(zk.DefaultZookeeperClient)}
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
		&handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper stateful-set secondary resources
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper service secondary resources
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}
	// Watch for changes to zookeeper pod secondary resources
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zookeeperv1beta1.ZookeeperCluster{},
	}, predicate.GenerationChangedPredicate{})
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
	client   client.Client
	scheme   *runtime.Scheme
	log      logr.Logger
	zkClient zk.ZookeeperClient
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
	if instance.GetTriggerRollingRestart() {
		r.log.Info("Restarting zookeeper cluster")
		annotationkey, annotationvalue := getRollingRestartAnnotation()
		if instance.Spec.Pod.Annotations == nil {
			instance.Spec.Pod.Annotations = make(map[string]string)
		}
		instance.Spec.Pod.Annotations[annotationkey] = annotationvalue
		instance.SetTriggerRollingRestart(false)
		changed = true
	}
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
		r.reconcileAdminServerService,
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

func getRollingRestartAnnotation() (string, string) {
	return "restartTime", time.Now().Format(time.RFC850)
}

// compareResourceVersion compare resoure versions for the supplied ZookeeperCluster and StatefulSet
// resources
// Returns:
//  0 if versions are equal
// -1 if ZookeeperCluster version is less than StatefulSet version
//  1 if ZookeeperCluster version is greater than StatefulSet version
func compareResourceVersion(zk *zookeeperv1beta1.ZookeeperCluster, sts *appsv1.StatefulSet) int {

	zkResourceVersion, zkErr := strconv.Atoi(zk.ResourceVersion)
	stsVersion, stsVersionFound := sts.Labels["owner-rv"]

	if !stsVersionFound {
		if zkErr != nil {
			log.Info("Fail to parse ZookeeperCluster version. Cannot decide zookeeper StatefulSet version")
			return 0
		}
		return 1
	}
	stsResourceVersion, err := strconv.Atoi(stsVersion)
	if err != nil {
		if zkErr != nil {
			log.Info("Fail to parse ZookeeperCluster version. Cannot decide zookeeper StatefulSet version")
			return 0
		}
		log.Info("Fail to convert StatefulSet version %s to integer; setting it to ZookeeperCluster version", stsVersion)
		return 1
	}
	if zkResourceVersion < stsResourceVersion {
		return -1
	} else if zkResourceVersion > stsResourceVersion {
		return 1
	}
	return 0
}

func (r *ReconcileZookeeperCluster) reconcileStatefulSet(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {

	// we cannot upgrade if cluster is in UpgradeFailed
	if instance.Status.IsClusterInUpgradeFailedState() {
		return nil
	}
	if instance.Spec.Pod.ServiceAccountName != "default" {
		serviceAccount := zk.MakeServiceAccount(instance)
		if err = controllerutil.SetControllerReference(instance, serviceAccount, r.scheme); err != nil {
			return err
		}
		// Check if this ServiceAccount already exists
		foundServiceAccount := &corev1.ServiceAccount{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccount.Name, Namespace: serviceAccount.Namespace}, foundServiceAccount)
		if err != nil && errors.IsNotFound(err) {
			r.log.Info("Creating a new ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
			err = r.client.Create(context.TODO(), serviceAccount)
			if err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			foundServiceAccount.ImagePullSecrets = serviceAccount.ImagePullSecrets
			r.log.Info("Updating ServiceAccount", "ServiceAccount.Namespace", serviceAccount.Namespace, "ServiceAccount.Name", serviceAccount.Name)
			err = r.client.Update(context.TODO(), foundServiceAccount)
			if err != nil {
				return err
			}
		}
	}
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
		// label the RV of the zookeeperCluster when creating the sts
		sts.Labels["owner-rv"] = instance.ResourceVersion
		err = r.client.Create(context.TODO(), sts)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		// check whether zookeeperCluster is updated before updating the sts
		cmp := compareResourceVersion(instance, foundSts)
		if cmp < 0 {
			return fmt.Errorf("Staleness: cr.ResourceVersion %s is smaller than labeledRV %s", instance.ResourceVersion, foundSts.Labels["owner-rv"])
		} else if cmp > 0 {
			// Zookeeper StatefulSet version inherits ZookeeperCluster resource version
			foundSts.Labels["owner-rv"] = instance.ResourceVersion
		}
		foundSTSSize := *foundSts.Spec.Replicas
		newSTSSize := *sts.Spec.Replicas
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
			r.zkClient.UpdateNode(path, data, version)
		}
		err = r.updateStatefulSet(instance, foundSts, sts)
		if err != nil {
			return err
		}
		return r.upgradeStatefulSet(instance, foundSts)
	}
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
	return nil
}

func (r *ReconcileZookeeperCluster) upgradeStatefulSet(instance *zookeeperv1beta1.ZookeeperCluster, foundSts *appsv1.StatefulSet) (err error) {

	//Getting the upgradeCondition from the zk clustercondition
	_, upgradeCondition := instance.Status.GetClusterCondition(zookeeperv1beta1.ClusterConditionUpgrading)

	if upgradeCondition == nil {
		// Initially set upgrading condition to false
		instance.Status.SetUpgradingConditionFalse()
		return nil
	}

	//Setting the upgrade condition to true to trigger the upgrade
	//When the zk cluster is upgrading Statefulset CurrentRevision and UpdateRevision are not equal and zk cluster image tag is not equal to CurrentVersion
	if upgradeCondition.Status == corev1.ConditionFalse {
		if instance.Status.IsClusterInReadyState() && foundSts.Status.CurrentRevision != foundSts.Status.UpdateRevision && instance.Spec.Image.Tag != instance.Status.CurrentVersion {
			instance.Status.TargetVersion = instance.Spec.Image.Tag
			instance.Status.SetPodsReadyConditionFalse()
			instance.Status.SetUpgradingConditionTrue("", "")
		}
	}

	//checking if the upgrade is in progress
	if upgradeCondition.Status == corev1.ConditionTrue {
		//checking when the targetversion is empty
		if instance.Status.TargetVersion == "" {
			r.log.Info("upgrading to an unknown version: cancelling upgrade process")
			return r.clearUpgradeStatus(instance)
		}
		//Checking for upgrade completion
		if foundSts.Status.CurrentRevision == foundSts.Status.UpdateRevision {
			instance.Status.CurrentVersion = instance.Status.TargetVersion
			r.log.Info("upgrade completed")
			return r.clearUpgradeStatus(instance)
		}
		//updating the upgradecondition if upgrade is in progress
		if foundSts.Status.CurrentRevision != foundSts.Status.UpdateRevision {
			r.log.Info("upgrade in progress")
			if fmt.Sprint(foundSts.Status.UpdatedReplicas) != upgradeCondition.Message {
				instance.Status.UpdateProgress(zookeeperv1beta1.UpdatingZookeeperReason, fmt.Sprint(foundSts.Status.UpdatedReplicas))
			} else {
				err = checkSyncTimeout(instance, zookeeperv1beta1.UpdatingZookeeperReason, foundSts.Status.UpdatedReplicas, 10*time.Minute)
				if err != nil {
					instance.Status.SetErrorConditionTrue("UpgradeFailed", err.Error())
					return r.client.Status().Update(context.TODO(), instance)
				} else {
					return nil
				}
			}
		}
	}
	return r.client.Status().Update(context.TODO(), instance)
}

func (r *ReconcileZookeeperCluster) clearUpgradeStatus(z *zookeeperv1beta1.ZookeeperCluster) (err error) {
	z.Status.SetUpgradingConditionFalse()
	z.Status.TargetVersion = ""
	// need to deep copy the status struct, otherwise it will be overwritten
	// when updating the CR below
	status := z.Status.DeepCopy()

	err = r.client.Update(context.TODO(), z)
	if err != nil {
		return err
	}

	z.Status = *status
	return nil
}

func checkSyncTimeout(z *zookeeperv1beta1.ZookeeperCluster, reason string, updatedReplicas int32, t time.Duration) error {
	lastCondition := z.Status.GetLastCondition()
	if lastCondition == nil {
		return nil
	}
	if lastCondition.Reason == reason && lastCondition.Message == fmt.Sprint(updatedReplicas) {
		// if reason and message are the same as before, which means there is no progress since the last reconciling,
		// then check if it reaches the timeout.
		parsedTime, _ := time.Parse(time.RFC3339, lastCondition.LastUpdateTime)
		if time.Now().After(parsedTime.Add(t)) {
			// timeout
			return fmt.Errorf("progress deadline exceeded")
		}
	}
	return nil
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

func (r *ReconcileZookeeperCluster) reconcileAdminServerService(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	svc := zk.MakeAdminServerService(instance)
	if err = controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return err
	}
	foundSvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}, foundSvc)
	if err != nil && errors.IsNotFound(err) {
		r.log.Info("Creating admin server service",
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
		r.log.Info("Updating existing admin server service",
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
	if instance.Status.IsClusterInUpgradingState() || instance.Status.IsClusterInUpgradeFailedState() {
		return nil
	}
	instance.Status.Init()
	foundPods := &corev1.PodList{}
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
		"StatefulSet.Namespace", instance.Namespace,
		"StatefulSet.Name", instance.Name)
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
	sts := zk.MakeStatefulSet(instance)

	subdir, err := yamlexporter.CreateOutputSubDir("ZookeeperCluster", sts.Labels["component"])
	return yamlexporter.GenerateOutputYAMLFile(subdir, sts.Kind, sts)
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

func (r *ReconcileZookeeperCluster) reconcileFinalizers(instance *zookeeperv1beta1.ZookeeperCluster) (err error) {
	if instance.Spec.Persistence != nil && instance.Spec.Persistence.VolumeReclaimPolicy != zookeeperv1beta1.VolumeReclaimPolicyDelete {
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
		MatchLabels: map[string]string{"app": instance.GetName(), "uid": string(instance.UID)},
	})
	pvclistOps := &client.ListOptions{
		Namespace:     instance.Namespace,
		LabelSelector: selector,
	}
	pvcList := &corev1.PersistentVolumeClaimList{}
	err = r.client.List(context.TODO(), pvcList, pvclistOps)
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
