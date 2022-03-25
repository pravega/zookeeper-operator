/**
 * Copyright (c) 2018-2022 Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2e

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Test Zookeeper cluster recreate with same name", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)
	podCount := 3

	Context("Create a default zookeeper cluster", func() {
		var (
			zk  *v1beta1.ZookeeperCluster
			err error
		)
		BeforeEach(func() {
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
		})
		It("should create the cluster successfully", func() {
			zk, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(zk).ToNot(BeNil())
			Expect(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: zk.Namespace, Name: zk.Name}, zk)
			}()).ToNot(HaveOccurred())
			// This checks that a reconcile event actually occurred since we can intercept them from requests
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			// A default Zookeeper cluster should have 3 replicas
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})

		It("Should be able to scale up the cluster", func() {
			// This is to get the latest zk cluster object
			zk, err = e2eutil.GetCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			// Scale up zk cluster, increase replicas to 5

			zk.Spec.Replicas = 5
			podCount = 5

			err = e2eutil.UpdateCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})

		It("should be able to scale down the cluster", func() {
			// This is to get the latest zk cluster object
			zk, err = e2eutil.GetCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			// Scale down zk cluster back to default
			zk.Spec.Replicas = 3
			podCount = 3

			err = e2eutil.UpdateCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})

		It("should successfully tear down the cluster", func() {
			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk), timeout).Should(Succeed())
		})
	})
})
