package stub

import (
	"context"
	"errors"
	"fmt"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
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
		o.WithDefaults()
		err := sdk.Create(newZkSts(o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper stateful-set : %v", err)
			return err
		}
		clientSvc, err := newZkClientSvc(o)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper client service : %v", err)
			return err
		}
		err = sdk.Create(clientSvc)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper client service : %v", err)
			return err
		}
		headlessSvc, err := newZkHeadlessSvc(o)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper headless service : %v", err)
			return err
		}
		err = sdk.Create(headlessSvc)
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper headless service : %v", err)
			return err
		}
	}
	return nil
}

// newZkSts creates a new Zookeeper StatefulSet
func newZkSts(z *v1beta1.ZookeeperCluster) *appsv1.StatefulSet {
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

// newZkClientSvc creates a new Zookeeper Service
func newZkClientSvc(z *v1beta1.ZookeeperCluster) (*v1.Service, error) {
	name := fmt.Sprintf("%s-client", z.GetObjectMeta().GetName())
	ports := []v1.ServicePort{}
	for _, p := range z.Spec.Ports {
		if p.Name == "client" {
			ports = append(ports, v1.ServicePort{Name: "client", Port: p.ContainerPort})
		} else {
			return nil, errors.New("No port named \"client\" in Cluster specification")
		}
	}
	return newSvc(name, ports, true, z), nil
}

// newZkInternalSvc creates a new Zookeeper Service
func newZkHeadlessSvc(z *v1beta1.ZookeeperCluster) (*v1.Service, error) {
	name := fmt.Sprintf("%s-headless", z.GetObjectMeta().GetName())
	ports := []v1.ServicePort{}
	for _, p := range z.Spec.Ports {
		if p.Name == "server" {
			ports = append(ports, v1.ServicePort{Name: "server", Port: p.ContainerPort})
		} else {
			return nil, errors.New("No port named \"server\" in Cluster specification")
		}
		if p.Name == "leader-election" {
			ports = append(ports, v1.ServicePort{Name: "leader-election", Port: p.ContainerPort})
		} else {
			return nil, errors.New("No port named \"leader-election\" in Cluster specification")
		}
	}
	return newSvc(name, ports, false, z), nil
}

func newSvc(name string, ports []v1.ServicePort, clusterIP bool, z *v1beta1.ZookeeperCluster) *v1.Service {
	service := v1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: z.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(z, schema.GroupVersionKind{
					Group:   v1beta1.SchemeGroupVersion.Group,
					Version: v1beta1.SchemeGroupVersion.Version,
					Kind:    "ZookeeperCluster",
				}),
			},
			Labels: map[string]string{
				"app": z.GetObjectMeta().GetName(),
			},
		},
		Spec: v1.ServiceSpec{Ports: ports},
	}
	if clusterIP == false {
		service.Spec.ClusterIP = v1.ClusterIPNone
	}
	return &service
}
