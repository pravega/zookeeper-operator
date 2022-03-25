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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Test multi-cluster zookeeper", func() {
	namespace := "default"
	podCount := 3

	Context("Create a default zookeeper cluster", func() {
		var (
			err            error
			zk1, zk2, zk3  *v1beta1.ZookeeperCluster
			initialVersion string
			upgradeVersion string
		)
		defaultCluster := e2eutil.NewDefaultCluster(namespace)
		defaultCluster.WithDefaults()
		defaultCluster.Status.Init()
		defaultCluster.ObjectMeta.Name = "zk1"
		defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

		It("should create the cluster successfully", func() {
			zk1, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default zookeeper cluster should have 3 replicas
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk1, podCount), timeout).Should(Succeed())
		})

		It("should create a second cluster successfully", func() {
			defaultCluster = e2eutil.NewDefaultCluster(namespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
			defaultCluster.ObjectMeta.Name = "zk2"
			initialVersion = "0.2.7"
			upgradeVersion = "0.2.9"
			defaultCluster.Spec.Image = v1beta1.ContainerImage{
				Repository: "pravega/zookeeper",
				Tag:        initialVersion,
			}
			zk2, err := e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk2, podCount), timeout).Should(Succeed())
		})

		It("should equal the initial version", func() {
			// This is to get the latest Zookeeper cluster object
			zk2, err = e2eutil.GetCluster(k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk2.Status.CurrentVersion).To(Equal(initialVersion))
			defaultCluster = e2eutil.NewDefaultCluster(namespace)
		})

		It("should create a 3rd cluster successfully", func() {
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.ObjectMeta.Name = "zk3"
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk3, err := e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk3, podCount), timeout).Should(Succeed())
		})

		It("should scale replicas of cluster 1 up & down", func() {
			// This is to get the latest zk cluster object
			zk1, err = e2eutil.GetCluster(k8sClient, zk1)
			// scale up the replicas in first cluster
			zk1.Spec.Replicas = 5
			podCount = 5

			err = e2eutil.UpdateCluster(k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk1, podCount), timeout).Should(Succeed())

			zk1, err = e2eutil.GetCluster(k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			// scale down the replicas back to 3
			zk1.Spec.Replicas = 3
			podCount = 3

			err = e2eutil.UpdateCluster(k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk1, podCount), timeout).Should(Succeed())
		})

		It("Should upgrade the image tag successfully in cluster 2", func() {
			// This is to get the latest Zookeeper cluster object
			zk2, err = e2eutil.GetCluster(k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			// upgrade the image in second Cluster
			zk2.Spec.Image.Tag = upgradeVersion

			err = e2eutil.UpdateCluster(k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToUpgrade(k8sClient, zk2, upgradeVersion), timeout).Should(Succeed())

			// This is to get the latest Zookeeper cluster object
			zk2, err = e2eutil.GetCluster(k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk2.Spec.Image.Tag).To(Equal(upgradeVersion))
			Expect(zk2.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(zk2.Status.TargetVersion).To(Equal(""))
		})

		It("should recreate pods for cluster 3", func() {
			// This is to get the latest zk cluster object
			zk3, err = e2eutil.GetCluster(k8sClient, zk3)
			// Delete all pods in the 3rd Cluster
			podDeleteCount := 3
			err = e2eutil.DeletePods(k8sClient, zk3, podDeleteCount)
			Expect(err).NotTo(HaveOccurred())
			time.Sleep(60 * time.Second)
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk3, podCount), timeout).Should(Succeed())
		})

		It("should delete all clusters", func() {
			err = e2eutil.DeleteCluster(k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk1), timeout).Should(Succeed())

			err = e2eutil.DeleteCluster(k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk2), timeout).Should(Succeed())

			err = e2eutil.DeleteCluster(k8sClient, zk3)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk3), timeout).Should(Succeed())
		})

		It("should recreate cluster 1 with the same name successfully", func() {
			// Recreating cluster with same name
			defaultCluster = e2eutil.NewDefaultCluster(namespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.ObjectMeta.Name = "zk1"
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk1, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk1, podCount), timeout).Should(Succeed())

			err = e2eutil.DeleteCluster(k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk1), timeout).Should(Succeed())
		})
	})
})
