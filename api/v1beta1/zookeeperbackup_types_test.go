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
	"github.com/pravega/zookeeper-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ZookeeperBackup Types", func() {
	var zkBk v1beta1.ZookeeperBackup
	BeforeEach(func() {
		zkBk = v1beta1.ZookeeperBackup{
			ObjectMeta: metav1.ObjectMeta{
				Name: "example",
			},
		}
	})

	Context("#WithDefaults", func() {
		var changed bool
		BeforeEach(func() {
			changed = zkBk.WithDefaults()
		})

		It("should return as changed", func() {
			Ω(changed).To(BeTrue())
		})

		It("should have a default schedule (every day)", func() {
			Ω(zkBk.Spec.Schedule).To(BeEquivalentTo("0 0 */1 * *"))
		})

		It("should have a default BackupsToKeep number", func() {
			Ω(zkBk.Spec.BackupsToKeep).To(BeEquivalentTo("7"))
		})

		It("should have a default DataCapacity size", func() {
			Ω(zkBk.Spec.DataCapacity).To(BeEquivalentTo("1Gi"))
		})

		It("should have a default image for backup", func() {
			Ω(zkBk.Spec.Image.Repository).To(BeEquivalentTo("pravega/zkbackup"))
			Ω(zkBk.Spec.Image.Tag).To(BeEquivalentTo("latest"))
		})
	})
})
