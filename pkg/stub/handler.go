package stub

import (
	"context"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewHandler() sdk.Handler {
	return &Handler{}
}

type Handler struct {
	// Fill me
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1beta1.ZookeeperCluster:
		err := sdk.Create(newZkSts(o))
		if err != nil && !errors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper stateful-set : %v", err)
			return err
		}
	}
	return nil
}

// newZkSts creates a new Zookeeper StatefulSet
func newZkSts(z *v1beta1.ZookeeperCluster) *appsv1.StatefulSet {
	z.WithDefaults()
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetObjectMeta().GetName(),
			Namespace: z.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(z, schema.GroupVersionKind{
					Group:   v1beta1.SchemeGroupVersion.Group,
					Version: v1beta1.SchemeGroupVersion.Version,
					Kind:    "ZookeeperCluster",
				}),
			},
			Labels: z.Spec.Pod.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &z.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": z.GetObjectMeta().GetName(),
				},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: z.GetObjectMeta().GetName(),
					Labels: map[string]string{
						"app": z.GetObjectMeta().GetName(),
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:            "zookeeper",
							Image:           z.Spec.Image.ToString(),
							ImagePullPolicy: z.Spec.Image.PullPolicy,
						},
					},
				},
			},
		},
	}
}
