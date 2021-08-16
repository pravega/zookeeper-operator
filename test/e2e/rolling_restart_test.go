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
	"time"
)

func testRollingRestart(t *testing.T) {
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

	zk, err := zk_e2eutil.CreateCluster(t, f, ctx, cluster)
	g.Expect(err).NotTo(HaveOccurred())

	// A default Zookeepercluster should have 3 replicas
	podSize := 3
	start := time.Now().Minute()*60 + time.Now().Second()
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	clusterCreateDuration := time.Now().Minute()*60 + time.Now().Second() - start
	g.Expect(err).NotTo(HaveOccurred())

	// This is to get the latest Zookeeper cluster object
	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())
	podList, err := zk_e2eutil.GetPods(t, f, zk)
	g.Expect(err).NotTo(HaveOccurred())
	for i := 0; i < len(podList.Items); i++ {
		g.Expect(podList.Items[i].Annotations).NotTo(HaveKey("restartTime"))
	}
	g.Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

	// Trigger a rolling restart
	zk.Spec.TriggerRollingRestart = true
	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk)
	// zk_e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
	time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())
	newPodList, err := zk_e2eutil.GetPods(t, f, zk)
	g.Expect(err).NotTo(HaveOccurred())
	var firstRestartTime []string
	for i := 0; i < len(newPodList.Items); i++ {
		g.Expect(newPodList.Items[i].Annotations).To(HaveKey("restartTime"))
		firstRestartTime = append(firstRestartTime, newPodList.Items[i].Annotations["restartTime"])
	}
	g.Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

	// Trigger a rolling restart again
	zk.Spec.TriggerRollingRestart = true
	err = zk_e2eutil.UpdateCluster(t, f, ctx, zk)
	// zk_e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the complete restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
	time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
	err = zk_e2eutil.WaitForClusterToBecomeReady(t, f, ctx, zk, podSize)
	g.Expect(err).NotTo(HaveOccurred())

	zk, err = zk_e2eutil.GetCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())
	newPodList2, err := zk_e2eutil.GetPods(t, f, zk)
	g.Expect(err).NotTo(HaveOccurred())
	for i := 0; i < len(newPodList2.Items); i++ {
		g.Expect(newPodList2.Items[i].Annotations).To(HaveKey("restartTime"))
		g.Expect(newPodList2.Items[i].Annotations["restartTime"]).NotTo(Equal(firstRestartTime[i]))
	}
	g.Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

	// Delete cluster
	err = zk_e2eutil.DeleteCluster(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

	// No need to do cleanup since the cluster CR has already been deleted
	doCleanup = false

	err = zk_e2eutil.WaitForClusterToTerminate(t, f, ctx, zk)
	g.Expect(err).NotTo(HaveOccurred())

}
