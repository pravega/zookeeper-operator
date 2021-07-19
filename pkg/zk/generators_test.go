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
	"fmt"

	log "github.com/sirupsen/logrus"

	"strings"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/utils"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generators Spec", func() {

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
					Spec: v1beta1.ZookeeperClusterSpec{
						Labels: map[string]string{
							"exampleLabel": "exampleValue",
						},
						Conf: v1beta1.ZookeeperConfig{
							AdditionalConfig: map[string]string{
								"tcpKeepAlive": "true",
							},
						},
					},
				}
				z.WithDefaults()
				cm = zk.MakeConfigMap(z)
			})

			It("should have custom labels set", func() {
				Ω(cm.GetLabels()).To(HaveKeyWithValue(
					"exampleLabel",
					"exampleValue"))
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

				It("should set additional configuration tcpKeepAlive to 'true'", func() {
					Ω(cfg).To(ContainSubstring("tcpKeepAlive=true\n"))
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

				It("should set the ADMIN_SERVER_HOST", func() {
					Ω(cfg).To(ContainSubstring("ADMIN_SERVER_HOST=example-admin-server\n"))
				})

				It("should set the ADMIN_SERVER_PORT", func() {
					Ω(cfg).To(ContainSubstring("ADMIN_SERVER_PORT=8080\n"))
				})

				It("should set the LEADER_PORT", func() {
					Ω(cfg).To(ContainSubstring("LEADER_PORT=3888\n"))
				})

			})
		})
		Context("with overridden kubernetes cluster domain", func() {
			var cfg string

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{
						KubernetesClusterDomain: "foo.bar",
					},
				}
				z.WithDefaults()
				cm = zk.MakeConfigMap(z)
			})

			Context("env.sh", func() {
				BeforeEach(func() {
					cfg = cm.Data["env.sh"]
				})

				It("should set the DOMAIN to the overridden headless domain", func() {
					Ω(cfg).
						To(ContainSubstring(
							"DOMAIN=example-headless.default.svc.foo.bar\n"))
				})

			})
		})
	})

	Context("#MakeStatefulSet", func() {
		var sts *appsv1.StatefulSet

		Context("with defaults", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{
						Labels: map[string]string{
							"exampleLabel": "exampleValue",
						},
					},
				}
				z.WithDefaults()
				sts = zk.MakeStatefulSet(z)
			})

			It("should have custom labels set", func() {
				Ω(sts.GetLabels()).To(HaveKeyWithValue(
					"exampleLabel",
					"exampleValue"))
			})

			It("should have custom labels set on pods", func() {
				Ω(sts.Spec.Template.ObjectMeta.Labels).To(HaveKeyWithValue(
					"exampleLabel",
					"exampleValue"))
			})
		})

		Context("with pod policy annotations", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{
						Pod: v1beta1.PodPolicy{
							Annotations: map[string]string{
								"example-annotation": "example-value",
							},
						},
					},
				}
				z.WithDefaults()
				sts = zk.MakeStatefulSet(z)
			})

			It("should have custom labels set on pods", func() {
				Ω(sts.Spec.Template.ObjectMeta.Annotations).To(HaveKeyWithValue(
					"example-annotation",
					"example-value"))
			})
		})
	})

	Context("#MakeStatefulSet with Ephemeral storage", func() {
		var sts *appsv1.StatefulSet

		Context("with defaults", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{},
				}
				z.Spec = v1beta1.ZookeeperClusterSpec{
					StorageType: "ephemeral",
				}
				z.WithDefaults()
				sts = zk.MakeStatefulSet(z)
			})
			It("Checking the sts spec contains volumesource as EmptyDir", func() {
				Ω(strings.ContainsAny(fmt.Sprintf("%v", sts.Spec.Template.Spec.Volumes), "EmptyDirVolumeSource")).Should(Equal(true))
			})
		})
	})
	Context("#MakeStatefulSet with non default service account", func() {
		var sts *appsv1.StatefulSet

		Context("with defaults", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{},
				}
				z.Spec.Pod.ServiceAccountName = "zookeeper"
				z.WithDefaults()
				zk.MakeServiceAccount(z)
				sts = zk.MakeStatefulSet(z)
			})
			It("Checking the sts service account", func() {
				Ω(sts.Spec.Template.Spec.ServiceAccountName).To(Equal("zookeeper"))
			})
		})
	})

	Context("#MakeStatefulSet with init containers", func() {
		var sts *appsv1.StatefulSet

		Context("with defaults", func() {

			BeforeEach(func() {
				z := &v1beta1.ZookeeperCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "example",
						Namespace: "default",
					},
					Spec: v1beta1.ZookeeperClusterSpec{},
				}
				z.WithDefaults()
				z.Spec.InitContainers = []v1.Container{
					v1.Container{
						Name:    "testing",
						Image:   "dummy-image",
						Command: []string{"sh", "-c", "ls;pwd"},
					},
				}
				sts = zk.MakeStatefulSet(z)
			})
			It("Checking the init containers", func() {
				log.Printf("init container is %v", sts.Spec.Template.Spec)
				Ω(sts.Spec.Template.Spec.InitContainers[0].Name).To(Equal("testing"))
				Ω(sts.Spec.Template.Spec.InitContainers[0].Image).To(Equal("dummy-image"))
			})
		})
	})

	Context("#MakeClientService", func() {
		var s *v1.Service
		var domainName string

		BeforeEach(func() {
			domainName = "zk.com."
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					DomainName: domainName,
					Labels: map[string]string{
						"exampleLabel": "exampleValue",
					},
					ClientService: v1beta1.ClientServicePolicy{
						Annotations: map[string]string{
							"exampleAnnotation": "exampleValue",
						},
					},
				},
			}
			z.WithDefaults()
			s = zk.MakeClientService(z)
		})

		It("should have a client port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-client")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(2181))
		})

		It("should have a client svc name", func() {
			Ω(s.GetName()).To(Equal("example-client"))
		})

		It("should have a client svc name", func() {
			Ω(s.Spec.Selector["app"]).To(Equal("example"))
		})

		It("should not set the dns annotation", func() {
			Expect(s.GetAnnotations()).NotTo(HaveKey("external-dns.alpha.kubernetes.io/hostname"))
		})

		It("should have custom labels set", func() {
			Ω(s.GetLabels()).To(HaveKeyWithValue(
				"exampleLabel",
				"exampleValue"))
		})

		It("should have custom annotations set", func() {
			Ω(s.GetAnnotations()).To(HaveKeyWithValue(
				"exampleAnnotation",
				"exampleValue"))
		})
	})

	Context("#MakeHeadlessService", func() {
		var s *v1.Service
		var domainName string

		BeforeEach(func() {
			domainName = "zk.com."
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					DomainName: domainName,
					Labels: map[string]string{
						"exampleLabel": "exampleValue",
					},
					HeadlessService: v1beta1.HeadlessServicePolicy{
						Annotations: map[string]string{
							"exampleAnnotation": "exampleValue",
						},
					},
				},
			}
			z.WithDefaults()
			s = zk.MakeHeadlessService(z)
		})

		It("should have a client port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-client")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(2181))
		})

		It("should have a quorum port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-quorum")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(2888))
		})

		It("should have a leader port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-leader-election")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(3888))
		})

		It("should have a metrics port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-metrics")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(7000))
		})

		It("should have a client svc name", func() {
			Ω(s.GetName()).To(Equal("example-headless"))
		})

		It("should have a client svc name", func() {
			Ω(s.Spec.Selector["app"]).To(Equal("example"))
		})

		It("should set the dns annotation", func() {
			Expect(s.GetAnnotations()).To(HaveKeyWithValue(
				"external-dns.alpha.kubernetes.io/hostname",
				"example-headless.zk.com."))
		})

		It("should have custom labels set", func() {
			Ω(s.GetLabels()).To(HaveKeyWithValue(
				"exampleLabel",
				"exampleValue"))
		})

		It("should have custom annotations set", func() {
			Ω(s.GetAnnotations()).To(HaveKeyWithValue(
				"exampleAnnotation",
				"exampleValue"))
		})
	})

	Context("#MakeHeadlessService dnsname without dot", func() {
		var s *v1.Service
		var domainName string

		BeforeEach(func() {
			domainName = "zkcom"
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					DomainName: domainName,
				},
			}
			z.WithDefaults()
			s = zk.MakeHeadlessService(z)
		})

		It("should set the dns annotation", func() {
			Expect(s.GetAnnotations()).To(HaveKeyWithValue(
				"external-dns.alpha.kubernetes.io/hostname",
				"example-headless.zkcom."))
		})
	})

	Context("#MakeAdminServerService", func() {
		var s *v1.Service

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					Labels: map[string]string{
						"exampleLabel": "exampleValue",
					},
					AdminServerService: v1beta1.AdminServerServicePolicy{
						Annotations: map[string]string{
							"exampleAnnotation": "exampleValue",
						},
					},
				},
			}
			z.WithDefaults()
			s = zk.MakeAdminServerService(z)
		})

		It("should have an admin server port", func() {
			p, err := utils.ServicePortByName(s.Spec.Ports, "tcp-admin-server")
			Ω(err).To(BeNil())
			Ω(p.Port).To(BeEquivalentTo(8080))
		})

		It("should have an admin server svc name", func() {
			Ω(s.GetName()).To(Equal("example-admin-server"))
		})

		It("should have a client svc name", func() {
			Ω(s.Spec.Selector["app"]).To(Equal("example"))
		})

		It("should have custom labels set", func() {
			Ω(s.GetLabels()).To(HaveKeyWithValue(
				"exampleLabel",
				"exampleValue"))
		})

		It("should have custom annotations set", func() {
			Ω(s.GetAnnotations()).To(HaveKeyWithValue(
				"exampleAnnotation",
				"exampleValue"))
		})

		It("should have no LoadBalancer attached by default", func() {
			Ω(s.Spec.Type).NotTo(Equal(v1.ServiceTypeLoadBalancer))
		})
	})

	Context("#MakeAdminServerService external with LoadBalancer", func() {
		var s *v1.Service

		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					AdminServerService: v1beta1.AdminServerServicePolicy{
						External: true,
					},
				},
			}
			z.WithDefaults()
			s = zk.MakeAdminServerService(z)
		})

		It("should have LoadBalancer attached", func() {
			Ω(s.Spec.Type).To(Equal(v1.ServiceTypeLoadBalancer))
		})
	})

	Context("#MakePodDisruptionBudget", func() {
		var pdb *policyv1beta1.PodDisruptionBudget
		var domainName string
		var zkClusterName string

		BeforeEach(func() {
			domainName = "zk.com."
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: v1beta1.ZookeeperClusterSpec{
					DomainName: domainName,
					Labels: map[string]string{
						"exampleLabel": "exampleValue",
					},
				},
			}
			z.WithDefaults()
			pdb = zk.MakePodDisruptionBudget(z)
			zkClusterName = z.GetName()
		})

		It("should have kind PodDisruptionBudget", func() {
			Ω(pdb.GetObjectKind().GroupVersionKind().Kind).To(Equal("PodDisruptionBudget"))
		})

		It("should have selector is zookeeper cluster name", func() {
			Ω(pdb.Spec.Selector.MatchLabels["app"]).To(BeEquivalentTo(zkClusterName))
		})

		It("should have custom labels set", func() {
			Ω(pdb.GetLabels()).To(HaveKeyWithValue(
				"exampleLabel",
				"exampleValue"))
		})
	})
})
