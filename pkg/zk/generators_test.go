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
	"testing"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/utils"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGenerators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Zk Spec")
}

var _ = Describe("Generators", func() {

	Context("#MakeConfigMap", func() {
		var cm *v1.ConfigMap

		Context("with defaults", func() {
			var cfg string

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
				}
				z.WithDefaults()
				cm = zk.MakeConfigMap(z)
			})

			Context("zoo.cfg", func() {
				BeforeEach(func() {
					cfg = cm.Data["zoo.cfg"]
				})

				It("should have a datadir of '/data'", func() {
					Ω(cfg).To(ContainSubstring("dataDir=/data\n"))
				})

				It("should set standaloneEnabled to 'false'", func() {
					Ω(cfg).To(ContainSubstring("standaloneEnabled=false\n"))
				})

				It("should set reconfigEnabled to 'true'", func() {
					Ω(cfg).To(ContainSubstring("reconfigEnabled=true\n"))
				})

				It("should set skipACL to 'yes'", func() {
					Ω(cfg).To(ContainSubstring("skipACL=yes"))
				})

				It("should set initLimit to '10'", func() {
					Ω(cfg).To(ContainSubstring("initLimit=10\n"))
				})

				It("should set tickTime to '2000'", func() {
					Ω(cfg).To(ContainSubstring("tickTime=2000\n"))
				})

				It("should set syncLimit to '2'", func() {
					Ω(cfg).To(ContainSubstring("syncLimit=2\n"))
				})

				It("should have a dynamicConfigFile", func() {
					Ω(cfg).
						To(ContainSubstring(
							"dynamicConfigFile=/data/zoo.cfg.dynamic\n"))
				})
			})

			Context("env.sh", func() {
				BeforeEach(func() {
					cfg = cm.Data["env.sh"]
				})

				It("should set the DOMAIN to the headless domain", func() {
					Ω(cfg).
						To(ContainSubstring(
							"DOMAIN=example-headless.default.svc.cluster.local\n"))
				})

				It("should set the QUORUM_PORT", func() {
					Ω(cfg).To(ContainSubstring("QUORUM_PORT=2888\n"))
				})

				It("should set the CLIENT_HOST", func() {
					Ω(cfg).To(ContainSubstring("CLIENT_HOST=example-client\n"))
				})

				It("should set the CLIENT_PORT", func() {
					Ω(cfg).To(ContainSubstring("CLIENT_PORT=2181\n"))
				})

				It("should set the LEADER_PORT", func() {
					Ω(cfg).To(ContainSubstring("LEADER_PORT=3888\n"))
				})

			})
		})
	})

	Context("#SyncStatefulSet", func() {
		var (
			sts1 *appsv1.StatefulSet
			err  error
		)

		Context("with a valid update specs", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
				}
				z.WithDefaults()
				sts1 = zk.MakeStatefulSet(z)
				sts2 := zk.MakeStatefulSet(z)
				reps := int32(4)
				sts2.Spec.Replicas = &reps
				sts2.Spec.Template.Spec.Containers[0].Image = "repo/newimage:latest"
				err = zk.SyncStatefulSet(sts1, sts2)
			})

			It("should not error", func() {
				Ω(err).To(BeNil())
			})

			It("should have the updated fields", func() {
				Ω(*sts1.Spec.Replicas).To(BeEquivalentTo(4))
				Ω(sts1.Spec.Template.Spec.Containers[0].Image).
					To(Equal("repo/newimage:latest"))
			})

		})

		Context("with invalid update specs", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
				}
				z.WithDefaults()
				sts1 := zk.MakeStatefulSet(z)
				sts2 := zk.MakeStatefulSet(z)
				sts2.Spec.ServiceName = "newservicename"
				err = zk.SyncStatefulSet(sts1, sts2)
			})

			It("should return an error", func() {
				Ω(err).NotTo(BeNil())
				Ω(err.Error()).To(Equal(zk.InvalidStatefulSetUpdateError))
			})
		})
	})

	Context("#MakeClientService", func() {
		var s *v1.Service

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			s = zk.MakeClientService(z)
		})

		It("should have a client port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "client")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(2181))
		})

		It("should have a the client svc name", func() {
			Ω(s.GetName()).To(Equal("example-client"))
		})

		It("should have a the client svc name", func() {
			Ω(s.Spec.Selector["app"]).To(Equal("example"))
		})
	})

	Context("#MakeHeadlessService", func() {
		var s *v1.Service

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			s = zk.MakeHeadlessService(z)
		})

		It("should have a quorum port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "quorum")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(2888))
		})

		It("should have a leader port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "leader-election")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(3888))
		})

		It("should have a the client svc name", func() {
			Ω(s.GetName()).To(Equal("example-headless"))
		})

		It("should have a the client svc name", func() {
			Ω(s.Spec.Selector["app"]).To(Equal("example"))
		})
	})
})
