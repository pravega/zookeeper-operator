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
	"github.com/pravega/zookeeper-operator/pkg/zk"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSynchronizers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Synchronizers Spec")
}

var _ = Describe("Synchronizers Spec", func() {

	Context("#SyncStatefulSet", func() {
		var sts1 *appsv1.StatefulSet

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
				zk.SyncStatefulSet(sts1, sts2)
			})

			It("should have the updated fields", func() {
				Ω(*sts1.Spec.Replicas).To(BeEquivalentTo(4))
				Ω(sts1.Spec.Template.Spec.Containers[0].Image).
					To(Equal("repo/newimage:latest"))
			})

		})

	})

})
