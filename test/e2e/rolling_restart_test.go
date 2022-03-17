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
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/test/e2e/e2eutil"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("Test rolling restart", func() {
	namespace := "default"
	defaultCluster := e2eutil.NewDefaultCluster(namespace)
	podCount := 3

	Context("Create a default zookeeper cluster", func() {
		var (
			zk                    *v1beta1.ZookeeperCluster
			err                   error
			clusterCreateDuration int
			firstRestartTime      []string
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
			start := time.Now().Minute()*60 + time.Now().Second()
			Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())
			clusterCreateDuration = time.Now().Minute()*60 + time.Now().Second() - start

		})

		It("should not have any restarts to begin with", func() {
			// This is to get the latest Zookeeper cluster object
			zk, err = e2eutil.GetCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			podList, err := e2eutil.GetPods(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())
			for i := 0; i < len(podList.Items); i++ {
				Expect(podList.Items[i].Annotations).NotTo(HaveKey("restartTime"))
			}
			Expect(zk.GetTriggerRollingRestart()).To(Equal(false))
		})

		Context("When triggering a rolling restart", func() {
			It("should update and become ready", func() {
				// Trigger a rolling restart
				zk.Spec.TriggerRollingRestart = true
				err = e2eutil.UpdateCluster(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				// e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
				time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
				Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())

				zk, err = e2eutil.GetCluster(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				newPodList, err := e2eutil.GetPods(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				for i := 0; i < len(newPodList.Items); i++ {
					Expect(newPodList.Items[i].Annotations).To(HaveKey("restartTime"))
					firstRestartTime = append(firstRestartTime, newPodList.Items[i].Annotations["restartTime"])
				}
				Expect(zk.GetTriggerRollingRestart()).To(Equal(false))
			})
			It("should update and become ready after 2nd rolling restart", func() {
				// Trigger a rolling restart again
				zk.Spec.TriggerRollingRestart = true
				err = e2eutil.UpdateCluster(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				// e2eutil.WaitForClusterToBecomeReady(...) will return as soon as any pod is restarted as the cluster is briefly reported to be healthy even though the complete restart is not completed. this method is hence called after a sleep to ensure that the restart has completed before asserting the test cases.
				time.Sleep(time.Duration(clusterCreateDuration) * 2 * time.Second)
				Eventually(e2eutil.WaitForClusterToBecomeReady(k8sClient, zk, podCount), timeout).Should(Succeed())

				zk, err = e2eutil.GetCluster(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				newPodList2, err := e2eutil.GetPods(k8sClient, zk)
				Expect(err).NotTo(HaveOccurred())
				for i := 0; i < len(newPodList2.Items); i++ {
					Expect(newPodList2.Items[i].Annotations).To(HaveKey("restartTime"))
					Expect(newPodList2.Items[i].Annotations["restartTime"]).NotTo(Equal(firstRestartTime[i]))
				}
				Expect(zk.GetTriggerRollingRestart()).To(Equal(false))
			})
		})
		It("should tear down the cluster", func() {
			// Delete cluster
			err = e2eutil.DeleteCluster(k8sClient, zk)
			Expect(err).NotTo(HaveOccurred())

			Eventually(e2eutil.WaitForClusterToTerminate(k8sClient, zk), timeout).Should(Succeed())
		})
	})
})
