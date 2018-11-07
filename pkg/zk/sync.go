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
	"fmt"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func syncClusterSize(zk *v1beta1.ZookeeperCluster) (err error) {
	sts := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      zk.GetName(),
			Namespace: zk.GetNamespace(),
		},
	}
	err = sdk.Get(sts)
	if err != nil {
		return fmt.Errorf("failed to get stateful-set (%s): %v", sts.Name, err)
	}

	zk.Status.Size = int(sts.Status.Replicas)
	err = sdk.Update(zk)
	if err != nil {
		return fmt.Errorf("failed to update project: %v", err)
	}

	if *sts.Spec.Replicas != zk.Spec.Size {
		sts.Spec.Replicas = &(zk.Spec.Size)
		err = sdk.Update(sts)
		if err != nil {
			return fmt.Errorf("failed to update size of stateful-set (%s): %v", sts.Name, err)
		}
	}

	return nil
}
