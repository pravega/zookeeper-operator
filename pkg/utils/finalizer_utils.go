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
	"strconv"
	"strings"
)

const (
	ZkFinalizer = "cleanUpZookeeperPVC"
)

func ContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}

func RemoveString(slice []string, str string) (result []string) {
	for _, item := range slice {
		if item == str {
			continue
		}
		result = append(result, item)
	}
	return result
}

func IsPVCOrphan(zkPvcName string, replicas int32) bool {
	index := strings.LastIndexAny(zkPvcName, "-")
	if index == -1 {
		return false
	}

	ordinal, err := strconv.Atoi(zkPvcName[index+1:])
	if err != nil {
		return false
	}

	return int32(ordinal) >= replicas
}
