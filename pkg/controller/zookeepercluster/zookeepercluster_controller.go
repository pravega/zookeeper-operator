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
	"time"

	"github.com/go-logr/logr"
	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	zookeeperv1beta1 "github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
	"github.com/q8s-io/zookeeper-operator-pravega/pkg/zk"
)

// ReconcileTime is the delay between reconciliations
const ReconcileTime = 10 * time.Second

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
	err = c.Watch(&source.Kind{Type: &statefulpodv1.StatefulPod{}}, &handler.EnqueueRequestForObject{})
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
	// err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
	err = c.Watch(&source.Kind{Type: &statefulpodv1.StatefulPod{}}, &handler.EnqueueRequestForOwner{
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
	client   client.Client
	scheme   *runtime.Scheme
	log      logr.Logger
	zkClient zk.ZookeeperClient
}

type reconcileFun func(cluster *zookeeperv1beta1.ZookeeperCluster) error

// Reconcile reads that state of the cluster for a ZookeeperCluster object and
// makes changes based on the state read and what is in the ZookeeperCluster.Spec
func (r *ReconcileZookeeperCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	r.log = log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
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
		r.reconcileConfigMap,
		r.reconcileStatefulPod,
		r.reconcileClusterStatus,
	} {
		if err = fun(instance); err != nil {
			return reconcile.Result{}, err
		}
	}
	// Recreate any missing resources every 'ReconcileTime'
	return reconcile.Result{RequeueAfter: ReconcileTime}, nil
}

// for test
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
