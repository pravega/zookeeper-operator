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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Zookeeper Finalizers", func() {

	Context("creating strSlice", func() {
		var val1, val2 bool
		BeforeEach(func() {
			var strSlice = []string{"10", "20"}
			val1 = ContainsString(strSlice, "10")
			val2 = ContainsString(strSlice, "30")
		})
		It("should return true for value 10", func() {
			Ω(val1).To(Equal(true))
		})
		It("should return false for value 30", func() {
			Ω(val2).To(Equal(false))
		})
	})

	Context("creating strSlice", func() {
		var result []string
		BeforeEach(func() {
			var strSlice = []string{"10", "20"}
			result = RemoveString(strSlice, "10")
		})
		It("should return false for value 10", func() {
			Ω(ContainsString(result, "10")).To(Equal(false))
		})
	})
	Context("IsPvCOrphan", func() {
		var result1, result2, result3, result4 bool
		BeforeEach(func() {
			var zkPvcName string = "zk"
			result1 = IsPVCOrphan(zkPvcName, 3)
			zkPvcName = "zk-2"
			result2 = IsPVCOrphan(zkPvcName, 3)
			zkPvcName = "zk-5"
			result3 = IsPVCOrphan(zkPvcName, 3)
			zkPvcName = "zk-"
			result4 = IsPVCOrphan(zkPvcName, 3)

		})
		It("should return false for result1", func() {
			Ω(result1).To(Equal(false))
		})
		It("should return false for result2", func() {
			Ω(result2).To(Equal(false))
		})
		It("should return true for result3", func() {
			Ω(result3).To(Equal(true))
		})
		It("should return false for result4", func() {
			Ω(result4).To(Equal(false))
		})
	})
})
