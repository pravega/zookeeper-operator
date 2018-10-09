/**
 * Copyright (c) 2018 Dell Inc., or its subsidiaries. All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 */

package stub

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"
	"github.com/pravega/zookeeper-operator/pkg/zk"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	if event.Deleted {
		// K8s will garbage collect and resources until zookeeper cluster delete
		return nil
	}

	switch o := event.Object.(type) {
	case *v1beta1.ZookeeperCluster:
		return zk.Reconcile(o)
	}
	return nil
}
