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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
)

var _ = Describe("Test Zookeeper cluster upgrade", func() {
	namespace := "default"
	podCount := 3

	Context("Create a default zookeeper cluster", func() {
		var (
			err            error
			zk             *v1beta1.ZookeeperCluster
			initialVersion string
			upgradeVersion string
		)
		defaultCluster := e2eutil.NewDefaultCluster(namespace)
		defaultCluster.WithDefaults()
		defaultCluster.Status.Init()
		defaultCluster.ObjectMeta.Name = "zk1"
		defaultCluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

		It("should create a cluster successfully", func() {
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
			zk, err = e2eutil.CreateCluster(k8sClient, defaultCluster)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
		})

		It("should upgrade the cluster", func() {
			// This is to get the latest Zookeeper cluster object
			zk, err = e2eutil.GetCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk.Status.CurrentVersion).To(Equal(initialVersion))

			zk.Spec.Image.Tag = upgradeVersion

			err = e2eutil.UpdateCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForClusterToUpgrade(k8sClient, zk, upgradeVersion)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should have the right versions", func() {

			// This is to get the latest Zookeeper cluster object
			zk, err = e2eutil.GetCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Expect(zk.Spec.Image.Tag).To(Equal(upgradeVersion))
			Expect(zk.Status.CurrentVersion).To(Equal(upgradeVersion))
			Expect(zk.Status.TargetVersion).To(Equal(""))
		})
		It("should tear down the cluster", func() {
			// Delete cluster
			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			err = e2eutil.WaitForClusterToTerminate(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
