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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and delete 1, 2 and all the pods
var _ = Describe("Test pods are recreated", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)
	podCount := 3

	Context("Create a default zookeeper cluster", func() {
		var (
			zk             *v1beta1.ZookeeperCluster
			err            error
			podDeleteCount int
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
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

			// A default zookeeper cluster should have 3 pods
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})

		It("should restart 1 pod", func() {
			podDeleteCount := 1
			err = e2eutil.DeletePods(k8sClient, zk, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})
		It("should restart when multiple pods are deleted", func() {
			podDeleteCount = 2
			err = e2eutil.DeletePods(k8sClient, zk, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(60 * time.Second)

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())

			podDeleteCount = 3
			err = e2eutil.DeletePods(k8sClient, zk, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(60 * time.Second)
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})
		It("should successfully tear down the cluster", func() {
			// Delete cluster
			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk), timeout).Should(Succeed())
		})
	})
})
