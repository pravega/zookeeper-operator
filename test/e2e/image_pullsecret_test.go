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
	v1 "k8s.io/api/core/v1"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Image pull secret check", func() {
	Context("Check create cluster with specified ImagePullSecrets ", func() {
		It("should create cluster with specified ImagePullSecrets specs", func() {

			defaultCluster := zk_e2eutil.NewDefaultCluster(testNamespace)
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
			By("create zk cluster with non-default spec")
			zk, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Zookeeper cluster should have 3 replicas
			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())
			Expect(zk_e2eutil.CheckAdminService(logger, k8sClient, zk)).NotTo(HaveOccurred())

			By("delete zk cluster")
			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())
			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())
		})
	})
})
