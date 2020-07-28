/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package e2eutil

import (
	api "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewDefaultCluster returns a cluster with an empty spec, which will be filled
// with default values
func NewDefaultCluster(namespace string) *api.ZookeeperCluster {
	return &api.ZookeeperCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ZookeeperCluster",
			APIVersion: "zookeeper.pravega.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "zookeeper",
			Namespace: namespace,
		},
		Spec: api.ZookeeperClusterSpec{},
	}
}

func NewClusterWithVersion(namespace, version string) *api.ZookeeperCluster {
	cluster := NewDefaultCluster(namespace)
	cluster.Spec = api.ZookeeperClusterSpec{
		Image: api.ContainerImage{
			Tag: version,
		},
	}
	return cluster
}

func NewClusterWithEmptyDir(namespace string) *api.ZookeeperCluster {
	cluster := NewDefaultCluster(namespace)
	cluster.Spec = api.ZookeeperClusterSpec{
		StorageType: "ephemeral",
	}
	return cluster
}
