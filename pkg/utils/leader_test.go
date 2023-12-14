/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package utils

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientscheme "k8s.io/client-go/kubernetes/scheme"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	configmapName  = "test-map"
	namespace      = "ns-1"
	currentPodName = "current-pod"
	otherPodName   = "some-other-pod"
)

var _ = Describe("Leader election utils", func() {
	Context("Election prechecks", func() {
		var (
			client        k8sClient.Client
			err           error
			ctx           context.Context
			lockConfigMap *corev1.ConfigMap
			currentPod    *corev1.Pod
			otherPod      *corev1.Pod
		)
		BeforeEach(func() {
			currentPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      currentPodName,
					UID:       "Uid-" + currentPodName,
					Namespace: namespace,
				},
			}
			otherPod = &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      otherPodName,
					UID:       "Uid-" + otherPodName,
					Namespace: namespace,
				},
			}
			_ = os.Setenv("POD_NAME", currentPodName)
			ctx = context.TODO()
		})

		When("leader lock owned by current pod", func() {
			BeforeEach(func() {
				lockConfigMap = &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      configmapName,
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Name: currentPodName, Kind: "Pod"},
						},
					},
				}

				client = fake.NewClientBuilder().WithScheme(clientscheme.Scheme).WithRuntimeObjects(
					[]runtime.Object{currentPod, otherPod, lockConfigMap}...).Build()

				err = precheckLeaderLock(ctx, client, configmapName, namespace)
			})
			It(" must do nothing", func() {
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("leader lock owned by other pod", func() {
			BeforeEach(func() {
				lockConfigMap = &corev1.ConfigMap{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ConfigMap",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      configmapName,
						Namespace: namespace,
						OwnerReferences: []metav1.OwnerReference{
							{Name: otherPodName, Kind: "Pod"},
						},
					},
				}
				client = fake.NewClientBuilder().WithScheme(clientscheme.Scheme).WithRuntimeObjects([]runtime.Object{currentPod, otherPod, lockConfigMap}...).Build()
				err = precheckLeaderLock(ctx, client, configmapName, namespace)
			})

			Context("when that node is Ready", func() {
				It(" must do nothing", func() {
					Expect(err).ShouldNot(HaveOccurred())

					pod := &corev1.Pod{}
					err = client.Get(ctx, k8sClient.ObjectKey{Namespace: namespace, Name: otherPodName}, pod)
					Expect(err).Should(BeNil())

					cm := &corev1.ConfigMap{}
					err = client.Get(ctx, k8sClient.ObjectKey{Namespace: namespace, Name: configmapName}, cm)
					Expect(err).Should(BeNil())
				})
			})

			Context("when that node is in ProviderFailed state", func() {
				BeforeEach(func() {
					otherPod.Status.Reason = "ProviderFailed"
					_ = client.Update(ctx, otherPod)

					err = precheckLeaderLock(ctx, client, configmapName, namespace)
				})
				It(" must delete otherPod and config map", func() {
					Expect(err).ShouldNot(HaveOccurred())

					pod := &corev1.Pod{}
					err = client.Get(ctx, k8sClient.ObjectKey{Namespace: namespace, Name: otherPodName}, pod)
					Expect(err).ShouldNot(BeNil())
					Expect(apierrors.IsNotFound(err)).To(BeTrue())

					cm := &corev1.ConfigMap{}
					err = client.Get(ctx, k8sClient.ObjectKey{Namespace: namespace, Name: configmapName}, cm)
					Expect(err).ShouldNot(BeNil())
					Expect(apierrors.IsNotFound(err)).To(BeTrue())
				})
			})
		})
	})
})
