package utils

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
)

// ServicePortByName returns a container port of name provided
func ServicePortByName(ports []corev1.ServicePort, name string) (port corev1.ServicePort, err error) {
	for _, port := range ports {
		if port.Name == name {
			return port, err
		}
	}
	return port, fmt.Errorf("port not found")
}
