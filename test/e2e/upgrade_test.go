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
	api "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

func testUpgradeCluster(t *testing.T) {
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

	cluster := zk_e2eutil.NewDefaultCluster(namespace)

	cluster.WithDefaults()
	cluster.Status.Init()
	cluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	initialVersion := "0.2.7"
	upgradeVersion := "0.2.9"
	cluster.Spec.Image = api.ContainerImage{
		Repository: "pravega/zookeeper",
		Tag:        initialVersion,
	}

	zk, err := zk_e2eutil.CreateCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Zookeepercluster should have 3 replicas
	podSize := 3
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Zookeeper cluster object
	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(zk.Status.CurrentVersion).To(Equal(initialVersion))

	zk.Spec.Image.Tag = upgradeVersion

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToUpgrade(t, f, ctx, zk, upgradeVersion)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Zookeeper cluster object
	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(zk.Spec.Image.Tag).To(Equal(upgradeVersion))
	g.Expect(zk.Status.CurrentVersion).To(Equal(upgradeVersion))
	g.Expect(zk.Status.TargetVersion).To(Equal(""))

	// Delete cluster
	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

}
