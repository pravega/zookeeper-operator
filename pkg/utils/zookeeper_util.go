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
	"strconv"
	"time"

	v1beta1 "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/samuel/go-zookeeper/zk"
	corev1 "k8s.io/api/core/v1"
)

const (
	// Set in https://github.com/pravega/bookkeeper/blob/master/docker/bookkeeper/entrypoint.sh#L21
	ZKMetaRoot = "/zookeeper-operator"
)

func GetZkConnection(zoo *v1beta1.ZookeeperCluster) (conn *zk.Conn, err error) {
	zkUri := GetZkServiceUri(zoo)
	host := []string{zkUri}
	conn, _, err = zk.Connect(host, time.Second*5)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to zookeeper: %v", err)
	}
	return conn, nil
}

func CreateZkMetaNode(zoo *v1beta1.ZookeeperCluster, conn *zk.Conn) (err error) {
	data := "CLUSTER_SIZE=" + strconv.Itoa(int(zoo.Spec.Replicas))
	if _, err := conn.Create(ZKMetaRoot, nil, 0, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("Error creating root zkNode: %v", err)
	}
	zNodePath := ZKMetaRoot + "/" + zoo.Name
	if _, err := conn.Create(zNodePath, []byte(data), 0, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("Error creating sub zkNode: %v", err)
	}
	return nil
}

func UpdateZkMetaNode(zoo *v1beta1.ZookeeperCluster, conn *zk.Conn, path string, data string, version int32) (err error) {
	if _, err := conn.Set(path, []byte(data), version); err != nil {
		return fmt.Errorf("Error updating zkNode: %v", err)
	}
	return nil
}

func GetZkServiceUri(zoo *v1beta1.ZookeeperCluster) (zkUri string) {
	zkClientPort, _ := ContainerPortByName(zoo.Spec.Ports, "client")
	zkUri = zoo.GetClientServiceName() + ":" + strconv.Itoa(int(zkClientPort))
	return zkUri
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

func GetMetaPath(zoo *v1beta1.ZookeeperCluster) (path string) {
	return fmt.Sprintf("%s/%s", ZKMetaRoot, zoo.Name)
}
