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
