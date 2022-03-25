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
var _ = Describe("Zookeeper cluster reconcile", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)
	podCount := 3

	BeforeEach(func() {
		defaultCluster.WithDefaults()
		defaultCluster.Status.Init()
		defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
	})

	Context("Creating a zookeeper cluster", func() {
		var (
			zk  *v1beta1.ZookeeperCluster
			err error
		)

		It("should reconcile after creating a zk cluster", func() {
			zk, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(zk).ToNot(BeNil())
			Expect(func() error {
				return k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: zk.Namespace, Name: zk.Name}, zk)
			}()).ToNot(HaveOccurred())
			Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
			// A default Zookeeper cluster should have 3 replicas
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})
		It("should have a service created", func() {
			err = e2eutil.CheckAdminService(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should delete and recreate successfully", func() {
			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForClusterToTerminate(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			zk, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForClusterToTerminate(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
