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
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Zookeeper Utils", func() {

	Context("with defaults", func() {
		var zkuri, path, containerport string
		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			zkuri = GetZkServiceUri(z)
			path = GetMetaPath(z)
			_, err := ContainerPortByName(z.Spec.Ports, "cl")
			if err != nil {
				containerport = err.Error()
			}
		})
		It("should set the zkuri", func() {
			Ω(zkuri).To(Equal("example-client.default.svc.cluster.local:2181"))
		})
		It("should set the path", func() {
			Ω(path).To(Equal("/zookeeper-operator/example"))
		})
		It("should give error message", func() {
			Ω(containerport).To(Equal("port not found"))
		})
	})
})
