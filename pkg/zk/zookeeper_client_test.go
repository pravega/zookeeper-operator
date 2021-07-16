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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Zookeeper Client", func() {

	Context("with a valid update of Service port", func() {
		var err1, err2, err3, err4, err5 error
		BeforeEach(func() {
			z := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			zkclient := new(zk.DefaultZookeeperClient)
			z.WithDefaults()
			err1 = zkclient.Connect("127.0.0.0:2181")
			err2 = zkclient.CreateNode(z, "temp/tmp/tmp")
			err5 = zkclient.CreateNode(z, "temp/tmp")
			err3 = zkclient.UpdateNode("temp/tem/temp", "dasd", 2)
			_, err4 = zkclient.NodeExists("temp")
			zkclient.Close()
		})
		It("err1 should be nil", func() {
			Ω(err1).Should(BeNil())
		})
		It("err2 should be not nil", func() {
			Ω(err2).ShouldNot(BeNil())
		})
		It("err3 should be not nil", func() {
			Ω(err3).ShouldNot(BeNil())
		})
		It("err4 should be not nil", func() {
			Ω(err4).ShouldNot(BeNil())
		})
		It("err5 should be not nil", func() {
			Ω(err5).ShouldNot(BeNil())
		})
	})
})
