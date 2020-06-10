/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package yamlexporter

import (
	"os"
	"testing"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestExporter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ZookeeperCluster yamlExporter")
}

var _ = Describe("ZookeeperCluster yamlExporter", func() {
	Context("with defaults", func() {
		var err, err2, err3, err4 error
		BeforeEach(func() {
			z1 := &v1beta1.ZookeeperCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
			}
			z1.WithDefaults()
			err = CreateYAMLOutputDir("test")
			_, err2 = ReadInputClusterYAMLFile("test")
			err3 = GenerateOutputYAMLFile("test", "test", z1.GetObjectMeta())
			_, err4 = CreateOutputSubDir(z1.GetName(), "test")
			_ = os.RemoveAll("test")
			_ = os.RemoveAll("example")
		})
		It("Err should be nil", func() {
			Ω(err).To(BeNil())
		})
		It("Err2 should give test: is a directory", func() {
			Ω(err2.Error()).To(Equal("read test: is a directory"))
		})
		It("Err3 should be nil", func() {
			Ω(err3).To(BeNil())
		})
		It("Err4 should be nil", func() {
			Ω(err4).To(BeNil())
		})
	})
})
