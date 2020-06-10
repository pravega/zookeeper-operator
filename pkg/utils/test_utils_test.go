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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Zookeeper Test_Utils", func() {
	Context("with defaults", func() {
		var port string
		var err string
		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z.WithDefaults()
			s := zk.MakeClientService(z)
			p, e := ServicePortByName(s.Spec.Ports, "temp")
			err = e.Error()
			p, e = ServicePortByName(s.Spec.Ports, "tcp-client")
			port = fmt.Sprintf("%v", p.Port)
		})
		It("should return error port not found for temp", func() {
			Ω(err).To(Equal("port not found"))
		})
		It("should set the serviceportbyname to 2181 for tcp-client", func() {
			Ω(port).To(Equal("2181"))
		})
	})
})
