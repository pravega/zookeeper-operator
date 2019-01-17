/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package zookeepercluster

import (
	"context"
	"testing"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestZookeepercluster(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ZookeeperCluster Controller Spec")
}

var _ = Describe("ZookeeperCluster Controller", func() {
	const (
		Name      = "example"
		Namespace = "default"
	)

	var (
		s = scheme.Scheme
		r *ReconcileZookeeperCluster
	)

	Context("Reconcile", func() {
		var (
			res reconcile.Result
			req reconcile.Request
			z   *v1beta1.ZookeeperCluster
		)

		BeforeEach(func() {
			req = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			z = &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      Name,
					Namespace: Namespace,
				},
			}
			s.AddKnownTypes(v1beta1.SchemeGroupVersion, z)
		})

		Context("Before defaults are applied", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				cl = fake.NewFakeClient(z)
				r = &ReconcileZookeeperCluster{client: cl, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("shouldn't error", func() {
				Ω(err).To(BeNil())
			})

			It("should set the default zk spec options", func() {
				foundZk := &v1beta1.ZookeeperCluster{}
				err = cl.Get(context.TODO(), req.NamespacedName, foundZk)
				Ω(err).To(BeNil())
				Ω(foundZk.Spec.Replicas).To(BeEquivalentTo(3))
			})

			It("should requeue the request", func() {
				Ω(res.Requeue).To(BeTrue())
			})

		})

		Context("After defaults are applied", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				z.WithDefaults()
				cl = fake.NewFakeClient(z)
				r = &ReconcileZookeeperCluster{client: cl, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("should not error", func() {
				Ω(err).To(BeNil())
			})

			It("should requeue after ReconfileTime delay", func() {
				Ω(res.RequeueAfter).To(Equal(ReconcileTime))
			})

			It("should create a config-map", func() {
				foundCm := &corev1.ConfigMap{}
				nn := types.NamespacedName{
					Name:      Name + "-configmap",
					Namespace: Namespace,
				}
				err = cl.Get(context.TODO(), nn, foundCm)
				Ω(err).To(BeNil())
			})

			It("should create a stateful-set", func() {
				foundSts := &appsv1.StatefulSet{}
				err = cl.Get(context.TODO(), req.NamespacedName, foundSts)
				Ω(err).To(BeNil())
				Ω(*foundSts.Spec.Replicas).To(BeEquivalentTo(3))
			})

			It("should create a client-service", func() {
				foundSvc := &corev1.Service{}
				nn := types.NamespacedName{
					Name:      Name + "-client",
					Namespace: Namespace,
				}
				err = cl.Get(context.TODO(), nn, foundSvc)
				Ω(err).To(BeNil())
			})

			It("should create a headless-service", func() {
				foundSvc := &corev1.Service{}
				nn := types.NamespacedName{
					Name:      Name + "-headless",
					Namespace: Namespace,
				}
				err = cl.Get(context.TODO(), nn, foundSvc)
				Ω(err).To(BeNil())
			})

			It("should create a pdb", func() {
				foundPdb := &policyv1beta1.PodDisruptionBudget{}
				err = cl.Get(context.TODO(), req.NamespacedName, foundPdb)
				Ω(err).To(BeNil())
			})

		})

		Context("With update to sts", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				z.WithDefaults()
				next := z.DeepCopy()
				st := zk.MakeStatefulSet(z)
				next.Spec.Replicas = 6
				cl = fake.NewFakeClient([]runtime.Object{next, st}...)
				r = &ReconcileZookeeperCluster{client: cl, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("should not raise an error", func() {
				Ω(err).To(BeNil())
			})

			It("should update the sts", func() {
				foundSts := &appsv1.StatefulSet{}
				err = cl.Get(context.TODO(), req.NamespacedName, foundSts)
				Ω(err).To(BeNil())
				Ω(*foundSts.Spec.Replicas).To(BeEquivalentTo(6))
			})
		})

		Context("With no update to sts", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				z.WithDefaults()
				next := z.DeepCopy()
				st := zk.MakeStatefulSet(z)
				cl = fake.NewFakeClient([]runtime.Object{next, st}...)
				r = &ReconcileZookeeperCluster{client: cl, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("should not raise an error", func() {
				Ω(err).To(BeNil())
			})

			It("should update the sts", func() {
				foundSts := &appsv1.StatefulSet{}
				err = cl.Get(context.TODO(), req.NamespacedName, foundSts)
				Ω(err).To(BeNil())
			})

		})

		Context("With an update to the client svc", func() {
			var (
				cl  client.Client
				err error
			)

			BeforeEach(func() {
				z.WithDefaults()
				next := z.DeepCopy()
				next.Spec.Ports[0].ContainerPort = 2182
				svc := zk.MakeClientService(z)
				cl = fake.NewFakeClient([]runtime.Object{next, svc}...)
				r = &ReconcileZookeeperCluster{client: cl, scheme: s}
				res, err = r.Reconcile(req)
			})

			It("should not raise an error", func() {
				Ω(err).ToNot(HaveOccurred())
			})
		})
	})
})
