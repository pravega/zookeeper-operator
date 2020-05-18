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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
)

var _ = Describe("ZookeeperCluster Status", func() {

	var z v1beta1.ZookeeperCluster

	BeforeEach(func() {
		z = v1beta1.ZookeeperCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
		}
	})

	Context("manually set pods ready condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionPodsReady,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			z.Status.Conditions = append(z.Status.Conditions, condition)
		})

		It("should contains pods ready condition and it is true status", func() {
			_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})

	Context("manually set pods upgrade condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionUpgrading,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			z.Status.Conditions = append(z.Status.Conditions, condition)
		})

		It("should contains pods upgrade condition and it is true status", func() {
			_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})
	Context("manually set pods Error condition to be true", func() {
		BeforeEach(func() {
			condition := v1beta1.ClusterCondition{
				Type:               v1beta1.ClusterConditionError,
				Status:             corev1.ConditionTrue,
				Reason:             "",
				Message:            "",
				LastUpdateTime:     "",
				LastTransitionTime: "",
			}
			z.Status.Conditions = append(z.Status.Conditions, condition)
		})

		It("should contains pods upgrade condition and it is true status", func() {
			_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionError)
			Ω(condition.Status).To(Equal(corev1.ConditionTrue))
		})
	})

	Context("set conditions for pods ready", func() {
		Context("set pods ready condition to be true", func() {
			BeforeEach(func() {
				z.Status.SetPodsReadyConditionFalse()
				z.Status.SetPodsReadyConditionTrue()
			})
			It("should have pods ready condition with true status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
		})

		Context("set pod ready condition to be false", func() {
			BeforeEach(func() {
				z.Status.SetPodsReadyConditionTrue()
				z.Status.SetPodsReadyConditionFalse()
			})

			It("should have ready condition with false status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should have updated timestamps", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionPodsReady)
				// TODO: check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
	})

	Context("set conditions for upgrade", func() {
		Context("set pods upgrade condition to be true", func() {
			BeforeEach(func() {
				z.Status.SetUpgradingConditionFalse()
				z.Status.SetUpgradingConditionTrue(" ", " ")
			})
			It("should have pods upgrade condition with true status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
		})

		Context("set pod upgrade condition to be false", func() {
			BeforeEach(func() {
				z.Status.SetUpgradingConditionTrue(" ", " ")
				z.Status.SetUpgradingConditionFalse()
			})

			It("should have upgrade condition with false status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should have updated timestamps", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionUpgrading)
				// TODO: check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
	})

	Context("set conditions for Error", func() {
		Context("set pods Error condition to be true", func() {
			BeforeEach(func() {
				z.Status.SetErrorConditionFalse()
				z.Status.SetErrorConditionTrue(" ", " ")
			})
			It("should have pods Error condition with true status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionError)
				Ω(condition.Status).To(Equal(corev1.ConditionTrue))
			})
		})

		Context("set pod Error condition to be false", func() {
			BeforeEach(func() {
				z.Status.SetErrorConditionTrue(" ", " ")
				z.Status.SetErrorConditionFalse()
			})

			It("should have Error condition with false status", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionError)
				Ω(condition.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should have updated timestamps", func() {
				_, condition := z.Status.GetClusterCondition(v1beta1.ClusterConditionError)
				// TODO: check the timestamps
				Ω(condition.LastUpdateTime).NotTo(Equal(""))
				Ω(condition.LastTransitionTime).NotTo(Equal(""))
			})
		})
	})

})
