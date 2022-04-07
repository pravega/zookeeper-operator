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
	api "github.com/pravega/zookeeper-operator/api/v1beta1"
	zk_e2eutil "github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
	"time"
)

// Test create and recreate a Zookeeper cluster with the same name
var _ = Describe("Operations with multiple cluster", func() {
	Context("Perform create, update, delete and recreate operations on 3 clusters", func() {
		It("should create cluster several clusters", func() {

			defaultCluster := zk_e2eutil.NewDefaultCluster(testNamespace)

			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.ObjectMeta.Name = "zk1"
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk1, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			// A default zookeeper cluster should have 3 replicas
			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk1, podSize)).NotTo(HaveOccurred())

			defaultCluster = zk_e2eutil.NewDefaultCluster(testNamespace)

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
			zk2, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk2, podSize)).NotTo(HaveOccurred())

			// This is to get the latest Zookeeper cluster object
			zk2, err = zk_e2eutil.GetCluster(logger, k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())
			Expect(zk2.Status.CurrentVersion).To(Equal(initialVersion))

			defaultCluster = zk_e2eutil.NewDefaultCluster(testNamespace)

			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.ObjectMeta.Name = "zk3"
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk3, err := zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk3, podSize)).NotTo(HaveOccurred())

			// This is to get the latest zk cluster object
			zk1, err = zk_e2eutil.GetCluster(logger, k8sClient, zk1)

			// scale up the replicas in first cluster
			zk1.Spec.Replicas = 5
			podSize = 5

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk1)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk1, podSize)).NotTo(HaveOccurred())

			zk1, err = zk_e2eutil.GetCluster(logger, k8sClient, zk1)
			Expect(err).NotTo(HaveOccurred())

			//scale down the replicas back to 3
			zk1.Spec.Replicas = 3
			podSize = 3

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk1)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk1, podSize)).NotTo(HaveOccurred())

			// This is to get the latest Zookeeper cluster object
			zk2, err = zk_e2eutil.GetCluster(logger, k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			//upgrade the image in second Cluster
			zk2.Spec.Image.Tag = upgradeVersion

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk2)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToUpgrade(logger, k8sClient, zk2, upgradeVersion)).NotTo(HaveOccurred())

			// This is to get the latest Zookeeper cluster object
			zk2, err = zk_e2eutil.GetCluster(logger, k8sClient, zk2)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk2.Spec.Image.Tag).To(Equal(upgradeVersion))
			Expect(zk2.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(zk2.Status.TargetVersion).To(Equal(""))

			// This is to get the latest zk cluster object
			zk3, err = zk_e2eutil.GetCluster(logger, k8sClient, zk3)

			//Delete all pods in the 3rd Cluster
			podDeleteCount := 3
			Expect(zk_e2eutil.DeletePods(logger, k8sClient, zk3, podDeleteCount)).NotTo(HaveOccurred())

			time.Sleep(60 * time.Second)
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk3, podSize)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk1)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk1)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk2)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk2)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk3)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk3)).NotTo(HaveOccurred())

			//Recreating cluster with same name
			defaultCluster = zk_e2eutil.NewDefaultCluster(testNamespace)
			defaultCluster.WithDefaults()
			defaultCluster.Status.Init()
			defaultCluster.ObjectMeta.Name = "zk1"
			defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk1, err = zk_e2eutil.CreateCluster(logger, k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk1, podSize)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk1)).NotTo(HaveOccurred())
			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk1)).NotTo(HaveOccurred())
		})
	})
})
