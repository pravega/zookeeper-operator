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

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeepcopy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ZookeeperCluster DeepCopy")
}

var _ = Describe("ZookeeperCluster DeepCopy", func() {
	Context("with defaults", func() {
		var str1, str2, str3, str4, str5, str6, str7, str8, str9 string
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
			//tempsts := z1.DeepCopyObject()
			str5 = z2.Spec.Image.Tag
			z2.Spec.Conf = *z1.Spec.Conf.DeepCopy()
			str6 = fmt.Sprintf("%v", z2.Spec.Conf.TickTime)
			z1.Spec.Conf.DeepCopyInto(&z2.Spec.Conf)
			str7 = fmt.Sprintf("%v", z2.Spec.Conf.TickTime)
			z2.Spec.Pod = *z1.Spec.Pod.DeepCopy()
			str8 = fmt.Sprintf("%v", z2.Spec.Pod.TerminationGracePeriodSeconds)
			z1.Spec.Pod.DeepCopyInto(&z2.Spec.Pod)
			str9 = fmt.Sprintf("%v", z2.Spec.Pod.TerminationGracePeriodSeconds)
			//z1.clus
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
		It("value of str4 image tag should be 0.2.6", func() {
			Ω(str4).To(Equal("0.2.6"))
		})
		It("value of str5 image tag should be 0.2.6", func() {
			Ω(str5).To(Equal("0.2.6"))
		})
		It("value of str6 image tag should be 0.2.6", func() {
			Ω(str6).To(Equal("2000"))
		})
		It("value of str7 image tag should be 0.2.6", func() {
			Ω(str7).To(Equal("2000"))
		})
		It("value of str8 image tag should be 0.2.6", func() {
			Ω(str8).To(Equal("30"))
		})
		It("value of str9 image tag should be 0.2.6", func() {
			Ω(str9).To(Equal("30"))
		})

	})
})
