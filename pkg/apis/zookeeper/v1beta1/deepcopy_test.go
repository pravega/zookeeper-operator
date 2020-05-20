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
		var str1, str2 string
		BeforeEach(func() {
			z1 := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z1.WithDefaults()
			z2 := *z1.DeepCopy()
			str1 = z1.Spec.Image.Tag
			str2 = z2.Spec.Image.Tag
		})
		It("should set the image for z2 as that of z1", func() {
			Ω(str2).To(Equal(str1))
		})
	})
})
