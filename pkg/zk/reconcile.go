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

import "github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"

func Reconcile(zk *v1beta1.ZookeeperCluster) (err error) {
	zk = zk.DeepCopy()
	zk.WithDefaults()

	deploy(zk)

	syncClusterSize(zk)

	return nil
}
