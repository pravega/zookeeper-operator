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

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
	v1 "k8s.io/api/core/v1"
)

// Test create and recreate a Zookeeper cluster with the same name
func testImagePullSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	doCleanup := true
	ctx := framework.NewTestCtx(t)
	defer func() {
		if doCleanup {
			ctx.Cleanup()
		}
	}()

	namespace, err := ctx.GetNamespace()
	g.Expect(err).NotTo(HaveOccurred())
	f := framework.Global

	defaultCluster := zk_e2eutil.NewDefaultCluster(namespace)

	defaultCluster.WithDefaults()
	defaultCluster.Status.Init()
	defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	defaultCluster.Spec.Image.Repository = "testanisha/zookeeper"
	defaultCluster.Spec.Image.Tag = "checksecret_1"
	defaultCluster.Spec.Pod.ImagePullSecrets = []v1.LocalObjectReference{
		{
			Name: "regcred",
		},
	}
	zk, err := zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Zookeeper cluster should have 3 replicas
	podSize := 3
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.CheckAdminService(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

}
