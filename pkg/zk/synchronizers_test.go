/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package zk_test

import (
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Synchronizers", func() {

	Context("with a valid update specs", func() {
		var sts1 *appsv1.StatefulSet

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			z.Spec.Pod.Resources.Limits = v1.ResourceList{
				v1.ResourceStorage: resource.MustParse("20Gi"),
			}
			sts1 = zk.MakeStatefulSet(z)
			sts2 := zk.MakeStatefulSet(z)
			reps := int32(4)
			sts2.Spec.Replicas = &reps
			sts2.Spec.Template.Spec.Containers[0].Image = "repo/newimage:latest"
			zk.SyncStatefulSet(sts1, sts2)
		})

		It("should have the updated fields", func() {
			Ω(*sts1.Spec.Replicas).To(BeEquivalentTo(4))
			Ω(sts1.Spec.Template.Spec.Containers[0].Image).
				To(Equal("repo/newimage:latest"))
		})
	})

	Context("with a valid update of Service port", func() {
		var port int32
		var value string

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			svc1 := zk.MakeClientService(z)
			svc2 := svc1.DeepCopy()
			svc2.Spec.Ports[0].Port = int32(4000)
			svc2.Spec.Type = "temp"
			zk.SyncService(svc1, svc2)
			port = svc1.Spec.Ports[0].Port
			value = string(svc1.Spec.Type)
		})

		It("should have the updated fields for service", func() {
			Ω(port).To(BeEquivalentTo(4000))
			Ω(value).To(BeEquivalentTo("temp"))
		})
	})

	Context("with a valid update of config map", func() {
		var value string
		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			cm1 := zk.MakeConfigMap(z)
			cm2 := cm1.DeepCopy()
			cm2.Data["k1"] = "v1"
			zk.SyncConfigMap(cm1, cm2)
			value = cm1.Data["k1"]
		})
		It("should have value as v1 for cm1.Data's key k1", func() {
			Ω(value).To(BeEquivalentTo("v1"))
		})
	})
})
