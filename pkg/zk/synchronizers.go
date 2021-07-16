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
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// SyncStatefulSet synchronizes any updates to the stateful-set
func SyncStatefulSet(curr *appsv1.StatefulSet, next *appsv1.StatefulSet) {
	curr.Spec.Replicas = next.Spec.Replicas
	curr.Spec.Template = next.Spec.Template
	curr.Spec.UpdateStrategy = next.Spec.UpdateStrategy
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
