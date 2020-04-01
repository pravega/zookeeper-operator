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
	v1beta1 "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

const (
	// Root ZNode for storing all zookeeper-operator related metadata.
	ZKMetaRoot = "/zookeeper-operator"
)

var (
	TopLevelDomain string
)

func GetZkServiceUri(zoo *v1beta1.ZookeeperCluster) (zkUri string) {
	zkClientPort, _ := ContainerPortByName(zoo.Spec.Ports, "client")
	zkUri = zoo.GetClientServiceName() + "." + zoo.GetNamespace() + ".svc.cluster." + TopLevelDomain + ":" + strconv.Itoa(int(zkClientPort))
	return zkUri
}

func GetMetaPath(zoo *v1beta1.ZookeeperCluster) (path string) {
	return fmt.Sprintf("%s/%s", ZKMetaRoot, zoo.Name)
}

// ContainerPortByName returns a container port of name provided
func ContainerPortByName(ports []corev1.ContainerPort, name string) (cPort int32, err error) {
	for _, port := range ports {
		if port.Name == name {
			return port.ContainerPort, nil
		}
	}
	return cPort, fmt.Errorf("port not found")
}
