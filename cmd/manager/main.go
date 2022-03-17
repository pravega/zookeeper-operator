/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
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
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/pravega/zookeeper-operator/pkg/apis"
	"github.com/pravega/zookeeper-operator/pkg/controller"
	zkConfig "github.com/pravega/zookeeper-operator/pkg/controller/config"
	"github.com/pravega/zookeeper-operator/pkg/utils"
	"github.com/pravega/zookeeper-operator/pkg/version"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	logz "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var (
	log         = logf.Log.WithName("cmd")
	versionFlag bool
)

func init() {
	flag.BoolVar(&versionFlag, "version", false, "Show version and quit")
	flag.BoolVar(&zkConfig.DisableFinalizer, "disableFinalizer", false,
		"Disable finalizers for zookeeperclusters. Use this flag with awareness of the consequences")
}

func printVersion() {
	log.Info(fmt.Sprintf("zookeeper-operator Version: %v", version.Version))
	log.Info(fmt.Sprintf("Git SHA: %s", version.GitSHA))
	log.Info(fmt.Sprintf("Go Version: %s", runtime.Version()))
	log.Info(fmt.Sprintf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH))
}

func main() {
	flag.Parse()

	// The logger instantiated here can be changed to any logger
	// implementing the logr.Logger interface. This logger will
	// be propagated through the whole operator, generating
	// uniform and structured logs.
	logf.SetLogger(logz.New(logz.UseDevMode(false)))

	printVersion()

	if versionFlag {
		os.Exit(0)
	}

	namespaces, ok := os.LookupEnv("WATCH_NAMESPACE")
	if !ok {
		log.Error(errors.New("env variable not set"), "failed to get watch namespace")
		os.Exit(1)
	}
	// When operator is started to watch resources in a specific set of namespaces, we use the MultiNamespacedCacheBuilder cache.
	// In this scenario, it is also suggested to restrict the provided authorization to this namespace by replacing the default
	// ClusterRole and ClusterRoleBinding to Role and RoleBinding respectively
	// For further information see the kubernetes documentation about
	// Using [RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).
	managerWatchCache := (cache.NewCacheFunc)(nil)
	if namespaces != "" {
		ns := strings.Split(namespaces, ",")
		for i := range ns {
			ns[i] = strings.TrimSpace(ns[i])
		}
		managerWatchCache = cache.MultiNamespacedCacheBuilder(ns)
	}

	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	operatorNs, err := GetOperatorNamespace()
	if err != nil {
		log.Error(err, "failed to get operator namespace")
		os.Exit(1)
	}

	// Become the leader before proceeding
	err = utils.BecomeLeader(context.TODO(), cfg, "zookeeper-operator-lock", operatorNs)
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{NewCache: managerWatchCache})
	if err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "")
		os.Exit(1)
	}

	log.Info("Starting the Cmd.")

	// Start the Cmd
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "manager exited non-zero")
		os.Exit(1)
	}
}

func GetOperatorNamespace() (string, error) {
	nsBytes, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		if os.IsNotExist(err) {
			return "", errors.New("file does not exist")
		}
		return "", err
	}
	ns := strings.TrimSpace(string(nsBytes))
	return ns, nil
}
