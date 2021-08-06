/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2e

import (
	"testing"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apis "github.com/pravega/zookeeper-operator/pkg/apis"
	operator "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestZookeeperCluster(t *testing.T) {
	zookeeperClusterList := &operator.ZookeeperClusterList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ZookeeperCluster",
			APIVersion: "zookeeper.pravega.io/v1beta1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, zookeeperClusterList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("x", testZookeeperCluster)
}

func testZookeeperCluster(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: zk_e2eutil.CleanupTimeout, RetryInterval: zk_e2eutil.CleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for zookeeper-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "zookeeper-operator", 1, zk_e2eutil.RetryInterval, zk_e2eutil.Timeout)
	if err != nil {
		t.Fatal(err)
	}

	testFuncs := map[string]func(t *testing.T){
		"testDeletePods":            testDeletePods,
		"testMultiZKCluster":        testMultiZKCluster,
		"testUpgradeCluster":        testUpgradeCluster,
		"testCreateRecreateCluster": testCreateRecreateCluster,
		"testImagePullSecret":       testImagePullSecret,
		"testScaleCluster":          testScaleCluster,
		"testEphemeralStorage":      testEphemeralStorage,
		"testRollingRestart":        testRollingRestart,
	}

	for name, f := range testFuncs {
		t.Run(name, f)
	}
}
