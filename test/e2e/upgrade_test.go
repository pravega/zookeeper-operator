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
)

var _ = Describe("Perform zk cluster upgrade", func() {
	Context("Check zk cluster upgrade operation", func() {
		It("should update spec image version", func() {

			cluster := zk_e2eutil.NewDefaultCluster(testNamespace)

			cluster.WithDefaults()
			cluster.Status.Init()
			cluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"
			initialVersion := "0.2.7"
			upgradeVersion := "0.2.9"
			cluster.Spec.Image = api.ContainerImage{
				Repository: "pravega/zookeeper",
				Tag:        initialVersion,
			}

			zk, err := zk_e2eutil.CreateCluster(logger, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Zookeepercluster should have 3 replicas
			podSize := 3
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			// This is to get the latest Zookeeper cluster object
			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk.Status.CurrentVersion).To(Equal(initialVersion))

			zk.Spec.Image.Tag = upgradeVersion

			Expect(zk_e2eutil.UpdateCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToUpgrade(logger, k8sClient, zk, upgradeVersion)).NotTo(HaveOccurred())

			// This is to get the latest Zookeeper cluster object
			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk.Spec.Image.Tag).To(Equal(upgradeVersion))
			Expect(zk.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(zk.Status.TargetVersion).To(Equal(""))

			// Delete cluster
			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())
		})
	})
})
