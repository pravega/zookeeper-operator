/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (&the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/operator-framework/operator-lib/leader"
	zkConfig "github.com/pravega/zookeeper-operator/pkg/controller/config"
	"github.com/pravega/zookeeper-operator/pkg/version"
	zkClient "github.com/pravega/zookeeper-operator/pkg/zk"
	"github.com/sirupsen/logrus"
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"os"
	"runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"strings"

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	api "github.com/pravega/zookeeper-operator/api/v1beta1"
	"github.com/pravega/zookeeper-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	log         = ctrl.Log.WithName("cmd")
	versionFlag bool
	scheme      = apimachineryruntime.NewScheme()
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
	flag.BoolVar(&zkConfig.DisableFinalizer, "disableFinalizer", false,
		"Disable finalizers for zookeeperclusters. Use this flag with awareness of the consequences")
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(api.AddToScheme(scheme))
}

func printVersion() {
	log.Info(fmt.Sprintf("zookeeper-operator Version: %v", version.Version))
	log.Info(fmt.Sprintf("Git SHA: %s", version.GitSHA))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	var metricsAddr string
	flag.StringVar(&metricsAddr, "metrics-bind-address", "127.0.0.1:6000", "The address the metric endpoint binds to.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	namespaces, err := getWatchNamespace()
	if err != nil {
		log.Error(err, "unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces")
	}
	printVersion()

	if versionFlag {
		os.Exit(0)
	}
	//When operator is started to watch resources in a specific set of namespaces, we use the MultiNamespacedCacheBuilder cache.
	//In this scenario, it is also suggested to restrict the provided authorization to this namespace by replacing the default
	//ClusterRole and ClusterRoleBinding to Role and RoleBinding respectively
	//For further information see the kubernetes documentation about
	//Using [RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).
	managerWatchCache := (cache.NewCacheFunc)(nil)
	if namespaces != "" {
		ns := strings.Split(namespaces, ",")
		for i := range ns {
			ns[i] = strings.TrimSpace(ns[i])
		}
		managerWatchCache = cache.MultiNamespacedCacheBuilder(ns)
	}
	ctx := context.TODO()

	// Become the leader before proceeding
	err = leader.Become(ctx, "zookeeper-operator-lock")
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               9443,
		NewCache:           managerWatchCache,
		Namespace:          namespaces,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		log.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.ZookeeperClusterReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("ZookeeperCluster"),
		Scheme:   mgr.GetScheme(),
		ZkClient: new(zkClient.DefaultZookeeperClient),
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "ZookeeperCluster")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	log.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() (string, error) {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}
	return ns, nil
}
