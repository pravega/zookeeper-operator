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
)

func testScaleCluster(t *testing.T) {
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

	zk, err := zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)

	g.Expect(err).NotTo(HaveOccurred())

	// A default zk cluster should have 3 pods
	podSize := 3
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest zk cluster object
	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale up zk cluster, increase replicas to 5

	zk.Spec.Replicas = 5
	podSize = 5

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	podDeleteCount := 2
	err = zk_e2eutil.DeletePods(t, f, ctx, zk, podDeleteCount)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest zk cluster object
	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	// Scale down zk cluster back to default
	zk.Spec.Replicas = 3
	podSize = 3

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// Delete cluster
	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

}
