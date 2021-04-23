/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package zk

import (
	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"
	"k8s.io/api/core/v1"
)

// SyncStatefulPod synchronizes any updates to the stateful-set
func SyncStatefulPod(curr *statefulpodv1.StatefulPod, next *statefulpodv1.StatefulPod) {
	curr.Spec.Size = next.Spec.Size
	curr.Spec.PodTemplate = next.Spec.PodTemplate
	// curr.Spec.UpdateStrategy = next.Spec.UpdateStrategy
	curr.Annotations = nil
}

// SyncService synchronizes a service with an updated spec and validates it
func SyncService(curr *v1.Service, next *v1.Service) {
	curr.Spec.Ports = next.Spec.Ports
	curr.Spec.Type = next.Spec.Type
}

// SyncConfigMap synchronizes a configmap with an updated spec and validates it
func SyncConfigMap(curr *v1.ConfigMap, next *v1.ConfigMap) {
	curr.Data = next.Data
	curr.BinaryData = next.BinaryData
}
