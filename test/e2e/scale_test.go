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

var _ = Describe("Perform scale for cluster upgrade", func() {
	Context("Check zk cluster scale operation", func() {
		It("should scale replicas number up and down", func() {
			defaultCluster := zk_e2eutil.NewDefaultCluster(testNamespace)

			defaultCluster.WithDefaults()

			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)

			Expect(err).NotTo(HaveOccurred())

			// A default zk cluster should have 3 pods
			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			// This is to get the latest zk cluster object
			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			// Scale up zk cluster, increase replicas to 5

			zk.Spec.Replicas = 5
			podSize = 5

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			podDeleteCount := 2
			Expect(zk_e2eutil.DeletePods(logger, k8sClient, zk, podDeleteCount)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			// This is to get the latest zk cluster object
			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			// Scale down zk cluster back to default
			zk.Spec.Replicas = 3
			podSize = 3

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			// Delete cluster
			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())
		})
	})
})
