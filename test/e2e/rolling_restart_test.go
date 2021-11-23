/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (&the "License");
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
	"time"
)

var _ = Describe("Perform rolling restart on zk cluster", func() {
	Context("Check rolling restart operation", func() {
		It("should perform rolling restart", func() {
			cluster := zk_e2eutil.NewDefaultCluster(testNamespace)

			cluster.WithDefaults()
			cluster.Status.Init()
			cluster.Spec.Persistence.VolumeReclaimPolicy = "Delete"

			zk, err := zk_e2eutil.CreateCluster(logger, k8sClient, cluster)
			Expect(err).NotTo(HaveOccurred())

			// A default Zookeepercluster should have 3 replicas
			podSize := 3
			start := time.Now().Minute()*60 + time.Now().Second()
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())
			clusterCreateDuration := time.Now().Minute()*60 + time.Now().Second() - start

			// This is to get the latest Zookeeper cluster object
			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			podList, err := zk_e2eutil.GetPods(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			for i := 0; i < len(podList.Items); i++ {
				Expect(podList.Items[i].Annotations).NotTo(HaveKey("restartTime"))
			}
			Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

			// Trigger a rolling restart
			zk.Spec.TriggerRollingRestart = true
			err = zk_e2eutil.UpdateCluster(logger, k8sClient, zk)
			// zk_e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
			time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			newPodList, err := zk_e2eutil.GetPods(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			var firstRestartTime []string
			for i := 0; i < len(newPodList.Items); i++ {
				Expect(newPodList.Items[i].Annotations).To(HaveKey("restartTime"))
				firstRestartTime = append(firstRestartTime, newPodList.Items[i].Annotations["restartTime"])
			}
			Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

			// Trigger a rolling restart again
			zk.Spec.TriggerRollingRestart = true
			err = zk_e2eutil.UpdateCluster(logger, k8sClient, zk)
			// zk_e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the complete restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
			time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
			Expect(zk_e2eutil.WaitForClusterToBecomeReady(logger, k8sClient, zk, podSize)).NotTo(HaveOccurred())

			zk, err = zk_e2eutil.GetCluster(logger, k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			newPodList2, err := zk_e2eutil.GetPods(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			for i := 0; i < len(newPodList2.Items); i++ {
				Expect(newPodList2.Items[i].Annotations).To(HaveKey("restartTime"))
				Expect(newPodList2.Items[i].Annotations["restartTime"]).NotTo(Equal(firstRestartTime[i]))
			}
			Expect(zk.GetTriggerRollingRestart()).To(Equal(false))

			// Delete cluster
			Expect(zk_e2eutil.DeleteCluster(logger, k8sClient, zk)).NotTo(HaveOccurred())

			Expect(zk_e2eutil.WaitForClusterToTerminate(logger, k8sClient, zk)).NotTo(HaveOccurred())
		})
	})
})
