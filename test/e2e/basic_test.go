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
	log "github.com/sirupsen/logrus"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Basic test controller", func() {
	Context("Check create/delete operations", func() {
		It("should create and recreate a Zookeeper cluster with the same name", func() {
			By("create Zookeeper cluster")
			log.Printf("--- 1 ---")
			defaultCluster := zk_e2eutil.NewDefaultCluster(testNamespace)
			log.Printf("--- 2 ---")
			defaultCluster.WithDefaults()
			log.Printf("--- 3 ---")
			defaultCluster.Status.Init()
			log.Printf("--- 4 ---")
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
			log.Printf("--- 5 ---")

			zk, err := zk_e2eutil.CreateCluster(&t, k8sClient, defaultCluster)
			log.Printf("--- 6 ---")
			Expect(err).NotTo(HaveOccurred())

			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(&t, k8sClient, defaultCluster, podSize)).NotTo(HaveOccurred())
			log.Printf("--- 7 ---")
			Expect(zk_e2eutil.CheckAdminService(&t, k8sClient, zk)).NotTo(HaveOccurred())
			log.Printf("--- 8 ---")

			By("delete created Zookeeper cluster")
			Expect(k8sClient.Delete(ctx, zk)).Should(Succeed())
			log.Printf("--- 9 ---")
			Expect(zk_e2eutil.WaitForClusterToTerminate(&t, k8sClient, zk)).NotTo(HaveOccurred())
			log.Printf("--- 10 ---")

			By("create Zookeeper cluster with the same name")
			defaultCluster = zk_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
			log.Printf("--- 11 ---")

			zk, err = zk_e2eutil.CreateCluster(&t, k8sClient, defaultCluster)
			log.Printf("--- 12 ---")
			Expect(err).NotTo(HaveOccurred())
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(&t, k8sClient, defaultCluster, podSize)).NotTo(HaveOccurred())
			log.Printf("--- 13 ---")

			By("delete created Zookeeper cluster")
			Expect(k8sClient.Delete(ctx, zk)).Should(Succeed())
			log.Printf("--- 14 ---")
			Expect(zk_e2eutil.WaitForClusterToTerminate(&t, k8sClient, zk)).NotTo(HaveOccurred())
			log.Printf("--- 15 ---")
		})
	})
})
