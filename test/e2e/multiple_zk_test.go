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
	"time"

	. "github.com/onsi/gomega"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	api "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Zookeeper cluster with the same name
func testMultiZKCluster(t *testing.T) {
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
	defaultCluster.ObjectMeta.Name = "zk1"
	defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

	zk1, err := zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default zookeeper cluster should have 3 replicas
	podSize := 3
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk1, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	defaultCluster = zk_e2eutil.NewDefaultCluster(namespace)

	defaultCluster.WithDefaults()
	defaultCluster.Status.Init()
	defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	defaultCluster.ObjectMeta.Name = "zk2"
	initialVersion := "0.2.7"
	upgradeVersion := "0.2.9"
	defaultCluster.Spec.Image = api.ContainerImage{
		Repository: "pravega/zookeeper",
		Tag:        initialVersion,
	}
	zk2, err := zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk2, podSize)
	g.Expect(err).NotTo(HaveOccurred())
	// This is to get the latest Zookeeper cluster object
	zk2, err = zk_e2eutil.GetCluster(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(zk2.Status.CurrentVersion).To(Equal(initialVersion))

	defaultCluster = zk_e2eutil.NewDefaultCluster(namespace)

	defaultCluster.WithDefaults()
	defaultCluster.Status.Init()
	defaultCluster.ObjectMeta.Name = "zk3"
	defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

	zk3, err := zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk3, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest zk cluster object
	zk1, err = zk_e2eutil.GetCluster(t, f, ctx, zk1)

	// scale up the replicas in first cluster
	zk1.Spec.Replicas = 5
	podSize = 5

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk1, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	zk1, err = zk_e2eutil.GetCluster(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	//scale down the replicas back to 3
	zk1.Spec.Replicas = 3
	podSize = 3

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk1, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Zookeeper cluster object
	zk2, err = zk_e2eutil.GetCluster(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	//upgrade the image in second Cluster
	zk2.Spec.Image.Tag = upgradeVersion

	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToUpgrade(t, f, ctx, zk2, upgradeVersion)
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Zookeeper cluster object
	zk2, err = zk_e2eutil.GetCluster(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(zk2.Spec.Image.Tag).To(Equal(upgradeVersion))
	g.Expect(zk2.Status.CurrentVersion).To(Equal(upgradeVersion))
	g.Expect(zk2.Status.TargetVersion).To(Equal(""))

	// This is to get the latest zk cluster object
	zk3, err = zk_e2eutil.GetCluster(t, f, ctx, zk3)

	//Delete all pods in the 3rd Cluster
	podDeleteCount := 3
	err = zk_e2eutil.DeletePods(t, f, ctx, zk3, podDeleteCount)
	g.Expect(err).NotTo(HaveOccurred())
	time.Sleep(60 * time.Second)
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk3, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk2)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk3)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk3)
	g.Expect(err).NotTo(HaveOccurred())

	//Recreating cluster with same name
	defaultCluster = zk_e2eutil.NewDefaultCluster(namespace)
	defaultCluster.WithDefaults()
	defaultCluster.Status.Init()
	defaultCluster.ObjectMeta.Name = "zk1"
	defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

	zk1, err = zk_e2eutil.CreateCluster(t, f, ctx, defaultCluster)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk1, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk1)
	g.Expect(err).NotTo(HaveOccurred())

}
