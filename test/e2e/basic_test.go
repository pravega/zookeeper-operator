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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Basic test controller", func() {
	Context("Check create/delete operations", func() {
		It("should create and recreate a Zookeeper cluster with the same name", func() {
			By("create Zookeeper cluster")
			defaultCluster := zk_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, defaultCluster, podSize)).NotTo(HaveOccurred())
			Expect(zk_e2eutil.CheckAdminService(logger, k8sClient, zk)).NotTo(HaveOccurred())

			By("delete created Zookeeper cluster")
			Expect(k8sClient.Delete(ctx, zk)).Should(Succeed())
			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())

			By("create Zookeeper cluster with the same name")
			defaultCluster = zk_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk, err = zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, defaultCluster, podSize)).NotTo(HaveOccurred())

			By("delete created Zookeeper cluster")
			Expect(k8sClient.Delete(ctx, zk)).Should(Succeed())
			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())
		})
	})
})
