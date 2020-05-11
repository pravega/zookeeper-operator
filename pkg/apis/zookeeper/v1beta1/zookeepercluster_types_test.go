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
	"testing"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestV1beta1(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ZookeeperCluster Types Spec")
}

var _ = Describe("ZookeeperCluster Types", func() {

	var z v1beta1.ZookeeperCluster

	BeforeEach(func() {
		z = v1beta1.ZookeeperCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "example",
			},
		}
	})

	Context("#WithDefaults", func() {
		var changed bool

		BeforeEach(func() {
			changed = z.WithDefaults()
		})

		It("should return as changed", func() {
			Ω(changed).To(BeTrue())
		})

		It("should have a replica count of 3", func() {
			Ω(z.Spec.Replicas).To(BeEquivalentTo(3))
		})

		It("should have an app label", func() {
			Ω(z.Spec.Labels["app"]).To(Equal("example"))
		})

		It("should have a release label", func() {
			Ω(z.Spec.Labels["release"]).To(Equal("example"))
		})

		Context("Image", func() {
			var i v1beta1.ContainerImage

			BeforeEach(func() {
				i = z.Spec.Image
			})

			It("should have the default repo", func() {
				Ω(i.Repository).To(Equal(v1beta1.DefaultZkContainerRepository))
			})

			It("should have the default tag", func() {
				Ω(i.Tag).To(Equal(v1beta1.DefaultZkContainerVersion))
			})

			It("should have the default policy", func() {
				Ω(i.PullPolicy).To(BeEquivalentTo(v1beta1.DefaultZkContainerPolicy))
			})
		})

		Context("Conf", func() {
			var c v1beta1.ZookeeperConfig

			BeforeEach(func() {
				c = z.Spec.Conf
			})

			It("should set InitLimit to 10", func() {
				Ω(c.InitLimit).To(Equal(10))
			})

			It("should set TickTime to 2000", func() {
				Ω(c.TickTime).To(Equal(2000))
			})

			It("should set SyncLimit to 2", func() {
				Ω(c.SyncLimit).To(Equal(2))
			})
		})

		Context("Ports", func() {
			var ports []corev1.ContainerPort

			BeforeEach(func() {
				ports = z.Spec.Ports
			})

			It("should have the default client port", func() {
				Ω(ports).To(ContainElement(corev1.ContainerPort{
					Name:          "client",
					ContainerPort: 2181,
				}))
			})

			It("should have the default quorum port", func() {
				Ω(ports).To(ContainElement(corev1.ContainerPort{
					Name:          "quorum",
					ContainerPort: 2888,
				}))
			})

			It("should have the default leader port", func() {
				Ω(ports).To(ContainElement(corev1.ContainerPort{
					Name:          "leader-election",
					ContainerPort: 3888,
				}))
			})

		})

		Context("Pod Policy", func() {
			var p v1beta1.PodPolicy

			BeforeEach(func() {
				p = z.Spec.Pod
			})

			It("should have an app label", func() {
				Ω(p.Labels["app"]).To(Equal("example"))
			})

			It("should have a release label", func() {
				Ω(p.Labels["release"]).To(Equal("example"))
			})

			Context("Pod Anti Affinity Rules", func() {
				var a corev1.WeightedPodAffinityTerm

				BeforeEach(func() {
					a = p.Affinity.PodAntiAffinity.
						PreferredDuringSchedulingIgnoredDuringExecution[0]
				})

				It("should have a weight of 20", func() {
					Ω(a.Weight).To(BeEquivalentTo(20))
				})

				It("should have a topology key of the hostname", func() {
					Ω(a.PodAffinityTerm.TopologyKey).To(Equal("kubernetes.io/hostname"))
				})

				It("should have a label selector based on the app", func() {
					Ω(a.PodAffinityTerm.LabelSelector.MatchExpressions).
						To(ContainElement(metav1.LabelSelectorRequirement{
							Key:      "app",
							Operator: metav1.LabelSelectorOpIn,
							Values:   []string{z.GetName()},
						}))
				})

			})
		})

		Context("PersistentVoluemClaim Spec", func() {
			var p corev1.PersistentVolumeClaimSpec

			BeforeEach(func() {
				p = z.Spec.Persistence.PersistentVolumeClaimSpec
			})

			It("should be an RWO volume", func() {
				Ω(p.AccessModes).To(Equal([]corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				}))
			})

			It("should have a 20Gi volume request", func() {
				Ω(p.Resources.Requests).To(Equal(corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("20Gi"),
				}))
			})
		})

	})

	Context("#ZookeeperPorts", func() {
		var p v1beta1.Ports

		BeforeEach(func() {
			z.WithDefaults()
			p = z.ZookeeperPorts()
		})

		It("should have a client port", func() {
			Ω(p.Client).To(BeEquivalentTo(2181))
		})

		It("should have a quorum port", func() {
			Ω(p.Quorum).To(BeEquivalentTo(2888))
		})

		It("should have a leader port", func() {
			Ω(p.Leader).To(BeEquivalentTo(3888))
		})

		It("should have a metrics port", func() {
			Ω(p.Metrics).To(BeEquivalentTo(7000))
		})
	})

})
