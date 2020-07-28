/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package v1beta1_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeepcopy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ZookeeperCluster DeepCopy")
}

var _ = Describe("ZookeeperCluster DeepCopy", func() {
	Context("with defaults", func() {
		var str1, str2, str3, str4, str5, str6, str7, str8, str9, str10 string
		var checkport int32
		var z2 *v1beta1.ZookeeperCluster
		BeforeEach(func() {
			z1 := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z1.WithDefaults()
			m := make(map[string]string)
			m["key"] = "value"
			z1.Annotations = m
			z1.Spec.Pod.NodeSelector = m
			temp := *z1.DeepCopy()
			z2 = &temp
			str1 = z1.Spec.Image.Tag
			str2 = z2.Spec.Image.Tag
			z1.Spec.Image.Tag = "0.2.5"
			z1.Spec.Image.DeepCopyInto(&z2.Spec.Image)
			str3 = z2.Spec.Image.Tag
			z1.Spec.Image.Tag = "0.2.6"
			z2.Spec.Image = *z1.Spec.Image.DeepCopy()
			str4 = z2.Spec.Image.Tag
			z2.Status = *z1.Status.DeepCopy()
			z2.Spec = *z1.Spec.DeepCopy()
			str5 = z2.Spec.Image.Tag
			z2.Spec.Conf = *z1.Spec.Conf.DeepCopy()
			str6 = fmt.Sprintf("%v", z2.Spec.Conf.TickTime)
			z1.Spec.Conf.DeepCopyInto(&z2.Spec.Conf)
			str7 = fmt.Sprintf("%v", z2.Spec.Conf.TickTime)
			z2.Spec.Pod = *z1.Spec.Pod.DeepCopy()
			str8 = fmt.Sprintf("%v", z2.Spec.Pod.TerminationGracePeriodSeconds)
			z1.Spec.Pod.DeepCopyInto(&z2.Spec.Pod)
			str9 = fmt.Sprintf("%v", z2.Spec.Pod.TerminationGracePeriodSeconds)
			p := z1.Spec.Ports[0].DeepCopy()
			z1.Spec.Ports[0].DeepCopyInto(&z2.Spec.Ports[0])
			z2.Spec.Ports[0].ContainerPort = p.ContainerPort
			z1.SetAnnotations(m)
			z2.Spec.Pod.Annotations = z1.Spec.Pod.Annotations
			z2.Spec.Persistence = z1.Spec.Persistence.DeepCopy()
			z1.Spec.Pod.DeepCopyInto(&z2.Spec.Pod)
			z1.Status.Members.Ready = []string{"zk-0", "zk-1"}
			z1.Status.Members.Unready = []string{"zk-2"}
			z2.Status.Members = *z1.Status.Members.DeepCopy()
			str10 = z2.Status.Members.Unready[0]
			env := &v1.EnvVar{}
			env.Name = "example"
			env.Value = "example-value"
			z1.Spec.Pod.Env = []v1.EnvVar{*env}
			tol := &v1.Toleration{}
			tol.Key = "tol"
			z1.Spec.Pod.Tolerations = []v1.Toleration{*tol}
			z1.Spec.Pod.Annotations = m
			podSec := &v1.PodSecurityContext{}
			z1.Spec.Pod.SecurityContext = podSec
			z1.Spec.Pod.DeepCopyInto(&z2.Spec.Pod)
			port := z1.ZookeeperPorts()
			tempPort := port.DeepCopy()
			checkport = tempPort.Client
		})
		It("value of str1 and str2 should be equal", func() {
			Ω(str2).To(Equal(str1))
		})
		It("value of str3 image tag should be 0.2.5", func() {
			Ω(str3).To(Equal("0.2.5"))
		})
		It("value of str4 image tag should be 0.2.6", func() {
			Ω(str4).To(Equal("0.2.6"))
		})
		It("value of str5 image tag should be 0.2.6", func() {
			Ω(str5).To(Equal("0.2.6"))
		})
		It("value of str6 should be 2000", func() {
			Ω(str6).To(Equal("2000"))
		})
		It("value of str7 should be 2000", func() {
			Ω(str7).To(Equal("2000"))
		})
		It("value of str8 should be 30", func() {
			Ω(str8).To(Equal("30"))
		})
		It("value of str9 should be 30", func() {
			Ω(str9).To(Equal("30"))
		})
		It("value of str10 should be zk-2", func() {
			Ω(str10).To(Equal("zk-2"))
		})

		It("value of checkport should be 2181", func() {
			Ω(checkport).To(Equal(int32(2181)))
		})

		It("value of env should be example", func() {
			Ω(z2.Spec.Pod.Env[0].Name).To(Equal("example"))
		})

		It("value of Tol should be example", func() {
			Ω(z2.Spec.Pod.Tolerations[0].Key).To(Equal("tol"))
		})

		It("checking for nil container image", func() {
			var contimage *v1beta1.ContainerImage
			contimage2 := contimage.DeepCopy()
			Ω(contimage2).To(BeNil())
		})
		It("checking for nil member status", func() {
			var memberstatus *v1beta1.MembersStatus
			memberstatus2 := memberstatus.DeepCopy()
			Ω(memberstatus2).To(BeNil())
		})
		It("checking for nil persistence", func() {
			var persistence *v1beta1.Persistence
			persistence2 := persistence.DeepCopy()
			Ω(persistence2).To(BeNil())
		})
		It("checking for nil clusterstatus", func() {
			var cluststatus *v1beta1.ZookeeperClusterStatus
			cluststatus2 := cluststatus.DeepCopy()
			Ω(cluststatus2).To(BeNil())
		})
		It("checking for nil clusterspec", func() {
			var clusterspec *v1beta1.ZookeeperClusterSpec
			clusterspec2 := clusterspec.DeepCopy()
			Ω(clusterspec2).To(BeNil())
		})
		It("checking for nil zookeeperconfig", func() {
			var zooconfig *v1beta1.ZookeeperConfig
			zooconfig2 := zooconfig.DeepCopy()
			Ω(zooconfig2).To(BeNil())
		})
		It("checking for nil clusterlist", func() {
			var clusterlist *v1beta1.ZookeeperClusterList
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2).To(BeNil())
		})
		It("checking for nil ports", func() {
			var ports1 *v1beta1.Ports
			ports2 := ports1.DeepCopy()
			Ω(ports2).To(BeNil())
		})
		It("checking for nil podpolicy", func() {
			var podpolicy *v1beta1.PodPolicy
			podpolicy2 := podpolicy.DeepCopy()
			Ω(podpolicy2).To(BeNil())
		})
		It("checking for nil zookeepercluster", func() {
			var zk *v1beta1.ZookeeperCluster
			zk2 := zk.DeepCopy()
			Ω(zk2).To(BeNil())
		})
		It("checking for deepcopy for clusterlist", func() {
			var clusterlist v1beta1.ZookeeperClusterList
			clusterlist.ResourceVersion = "v1beta1"
			clusterlist2 := clusterlist.DeepCopy()
			Ω(clusterlist2.ResourceVersion).To(Equal("v1beta1"))
		})
		It("checking for Deepcopy object", func() {
			zk := z2.DeepCopyObject()
			Ω(zk.GetObjectKind().GroupVersionKind().Version).To(Equal(""))
		})
		It("checking for deepcopyobject for clusterlist", func() {
			var clusterlist v1beta1.ZookeeperClusterList
			clusterlist.ResourceVersion = "v1beta1"
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
		It("checking for deepcopyobject for clusterlist with items", func() {
			var clusterlist v1beta1.ZookeeperClusterList
			clusterlist.ResourceVersion = "v1beta1"
			clusterlist.Items = []v1beta1.ZookeeperCluster{
				{
					Spec: v1beta1.ZookeeperClusterSpec{},
				},
			}
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).ShouldNot(BeNil())
		})
		It("checking for nil pravega cluster deepcopyobject", func() {
			var cluster *v1beta1.ZookeeperCluster
			cluster2 := cluster.DeepCopyObject()
			Ω(cluster2).To(BeNil())
		})
		It("checking for nil pravega clusterlist deepcopyobject", func() {
			var clusterlist *v1beta1.ZookeeperClusterList
			clusterlist2 := clusterlist.DeepCopyObject()
			Ω(clusterlist2).To(BeNil())
		})
	})
})
