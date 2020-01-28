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
	ZKMetaRoot = "/zookeeper-operator/"
	//ZkFinalizer = "cleanUpZookeeper"
)

func GetZkConnection(zoo *v1beta1.ZookeeperCluster) (conn *zk.Conn, err error) {
	zkUri := getZkServiceUri(zoo)
	host := []string{zkUri}
	conn, _, err = zk.Connect(host, time.Second*5)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to zookeeper: %v", err)
	}
	return conn, nil
}

func ExistsZkMetaRoot(zoo *v1beta1.ZookeeperCluster, conn *zk.Conn) (exists bool, err error) {
	if conn, err = GetZkConnection(zoo); err != nil {
		return exists, err
	}
	defer conn.Close()
	path := getMetaPath(zoo)
	pathexist, _, err := conn.Exists(path)
	if err != nil {
		return false, fmt.Errorf("failed to check if zookeeper path exists: %v", err)
	}

	return pathexist, nil
}

func CreateZkMetaNode(zoo *v1beta1.ZookeeperCluster, conn *zk.Conn) (err error) {
	if conn, err = GetZkConnection(zoo); err != nil {
		return err
	}
	defer conn.Close()
	data := "CLUSTER_SIZE=" + strconv.Itoa(int(zoo.Spec.Replicas))
	path := getMetaPath(zoo)
	if _, err := conn.Create(path, []byte(data), 0, zk.WorldACL(zk.PermAll)); err != nil {
		return fmt.Errorf("Error creating zkNode: %v", err)
	}
	return nil
}

func UpdateZkMetaNode(zoo *v1beta1.ZookeeperCluster, conn *zk.Conn, newSize int32) (err error) {
	if conn, err = GetZkConnection(zoo); err != nil {
		return err
	}
	defer conn.Close()
	data := "CLUSTER_SIZE=" + strconv.Itoa(int(newSize))
	path := getMetaPath(zoo)
	zoo.Status.MetadataVersion++
	if _, err := conn.Set(path, []byte(data), zoo.Status.MetadataVersion); err != nil {
		return fmt.Errorf("Error creating zkNode: %v", err)
	}
	return nil
}

/*
func ReadZkMetaNode(zoo *v1beta1.ZookeeperCluster) (err error) {

}
*/
// Delete all znodes related to a specific Bookkeeper cluster
/*
func DeleteAllZnodes(zoo *v1beta1.ZookeeperCluster) (err error) {
	zkUri := GetZkServiceUri(zoo)
	host := []string{zkUri}
	conn, _, err := zk.Connect(host, time.Second*5)
	if err != nil {
		return fmt.Errorf("failed to connect to zookeeper: %v", err)
	}
	defer conn.Close()

	root := getMetaPath(zoo)
	exist, _, err := conn.Exists(root)
	if err != nil {
		return fmt.Errorf("failed to check if zookeeper path exists: %v", err)
	}
	if exist {
		// Construct BFS tree to delete all znodes recursivelyreturn nil
	}
	tree, err := ListSubTreeBFS(conn, root)
	if err != nil {
		return fmt.Errorf("failed to construct BFS tree: %v", err)
	}

	for tree.Len() != 0 {
		err := conn.Delete(tree.Back().Value.(string), -1)
		if err != nil {
			return fmt.Errorf("failed to delete znode (%s): %v", tree.Back().Value.(string), err)
		}
		tree.Remove(tree.Back())
	}
	return nil
}
*/
// Construct a BFS tree
/*
func ListSubTreeBFS(conn *zk.Conn, root string) (*list.List, error) {
	queue := list.New()
	tree := list.New()
	queue.PushBack(root)
	tree.PushBack(root)

	for {
		if queue.Len() == 0 {
			break
		}
		node := queue.Front()
		children, _, err := conn.Children(node.Value.(string))
		if err != nil {
			return tree, err
		}

		for _, child := range children {
			childPath := fmt.Sprintf("%s/%s", node.Value.(string), child)
			queue.PushBack(childPath)
			tree.PushBack(childPath)
		}
		queue.Remove(node)
	}
	return tree, nil
}
*/
func getZkServiceUri(zoo *v1beta1.ZookeeperCluster) (zkUri string) {
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

func getMetaPath(zoo *v1beta1.ZookeeperCluster) (root string) {
	return fmt.Sprintf("/%s/%s/cluster", ZKMetaRoot, zoo.Name)
}
