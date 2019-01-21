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

	corev1 "k8s.io/api/core/v1"
)

// ServicePortByName returns a container port of name provided
func ServicePortByName(ports []corev1.ServicePort, name string) (port corev1.ServicePort, err error) {
	for _, port := range ports {
		if port.Name == name {
			return port, nil
		}
	}
	return port, fmt.Errorf("port not found")
}
