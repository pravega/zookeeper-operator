package stub

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/pravega/zookeeper-operator/pkg/apis/zookeeper/v1beta1"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
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
		configMapName := fmt.Sprintf("%s-configmap", o.GetName())
		err := sdk.Create(newZkConfigMap(configMapName, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper configmap : %v", err)
			return err
		}
		err = sdk.Create(newZkSts(configMapName, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper statefulset : %v", err)
			return err
		}
		err = sdk.Create(newZkPdb(o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper pod-disruption-budget : %v", err)
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
func newZkSts(configMapName string, z *v1beta1.ZookeeperCluster) *appsv1.StatefulSet {
	sts := appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName(),
			Namespace: z.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(z, schema.GroupVersionKind{
					Group:   v1beta1.SchemeGroupVersion.Group,
					Version: v1beta1.SchemeGroupVersion.Version,
					Kind:    "ZookeeperCluster",
				}),
			},
			Labels: z.Spec.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &z.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": z.GetName(),
				},
			},
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: z.GetName(),
					Labels: map[string]string{
						"app": z.GetName(),
					},
				},
				Spec: newZkPodSpec(configMapName, z),
			},
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:   "data",
						Labels: map[string]string{"app": z.GetName()},
					},
					Spec: *z.Spec.PersistentVolumeClaimSpec,
				},
			},
		},
	}
	return &sts
}

func newZkPodSpec(configMapName string, z *v1beta1.ZookeeperCluster) v1.PodSpec {
	container := v1.Container{
		Name:            "zookeeper",
		Image:           z.Spec.Image.ToString(),
		ImagePullPolicy: z.Spec.Image.PullPolicy,
		VolumeMounts: []v1.VolumeMount{
			{Name: "data", MountPath: "/data"},
			{Name: "conf", MountPath: "/conf"},
		},
	}
	if z.Spec.Pod.Resources.Limits != nil || z.Spec.Pod.Resources.Requests != nil {
		container.Resources = z.Spec.Pod.Resources
	}
	container.Env = z.Spec.Pod.Env
	podSpec := v1.PodSpec{
		Containers: []v1.Container{container},
		Affinity:   z.Spec.Pod.Affinity,
		Volumes: []v1.Volume{
			{
				Name: "conf",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{Name: configMapName},
					},
				},
			},
		},
		TerminationGracePeriodSeconds: &z.Spec.Pod.TerminationGracePeriodSeconds,
	}
	if reflect.DeepEqual(v1.PodSecurityContext{}, z.Spec.Pod.SecurityContext) {
		podSpec.SecurityContext = z.Spec.Pod.SecurityContext
	}
	podSpec.NodeSelector = z.Spec.Pod.NodeSelector
	podSpec.Tolerations = z.Spec.Pod.Tolerations

	return podSpec
}

// newZkClientSvc creates a new Zookeeper Service
func newZkClientSvc(z *v1beta1.ZookeeperCluster) (*v1.Service, error) {
	name := fmt.Sprintf("%s-client", z.GetName())
	ports := []v1.ServicePort{}
	for _, p := range z.Spec.Ports {
		if p.Name == "client" {
			ports = append(ports, v1.ServicePort{Name: "client", Port: p.ContainerPort})
		}
	}
	if len(ports) == 0 {
		return nil, errors.New("No port named \"client\" in Cluster specification")
	}
	return newSvc(name, ports, true, z), nil
}

// newZkInternalSvc creates a new Zookeeper Service
func newZkHeadlessSvc(z *v1beta1.ZookeeperCluster) (*v1.Service, error) {
	name := fmt.Sprintf("%s-headless", z.GetName())
	ports := []v1.ServicePort{}
	for _, p := range z.Spec.Ports {
		if p.Name == "server" {
			ports = append(ports, v1.ServicePort{Name: "server", Port: p.ContainerPort})
		}
		if p.Name == "leader-election" {
			ports = append(ports, v1.ServicePort{Name: "leader-election", Port: p.ContainerPort})
		}
	}
	if len(ports) < 2 {
		return nil, errors.New("You must provide a port named \"leader-election\" and \"server\" in Cluster specification")
	}
	return newSvc(name, ports, false, z), nil
}

func newZkConfigMap(name string, z *v1beta1.ZookeeperCluster) *v1.ConfigMap {
	return &v1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
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
		},
		Data: map[string]string{
			"zoo.cfg":          newZkConfigString(z),
			"log4j.properties": newZkLog4JConfigString(),
		},
	}
}

func newZkConfigString(z *v1beta1.ZookeeperCluster) string {
	var b strings.Builder
	b.WriteString(
		"4lw.commands.whitelist=mntr, ruok\n" +
			"dataDir=/data\n" +
			"standaloneEnabled=false\n" +
			"reconfigEnabled=true\n" +
			"dynamicConfigFile=/conf/zoo.cfg.dynamic\n",
	)
	return b.String()
}

func newZkLog4JConfigString() string {
	return "zookeeper.root.logger=CONSOLE\n" +
		"zookeeper.console.threshold=INFO\n" +
		"log4j.rootLogger=${zookeeper.root.logger}\n" +
		"log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n" +
		"log4j.appender.CONSOLE.Threshold=${zookeeper.console.threshold}\n" +
		"log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n" +
		"log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n\n"
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
				"app": z.GetName(),
			},
		},
		Spec: v1.ServiceSpec{Ports: ports},
	}
	if clusterIP == false {
		service.Spec.ClusterIP = v1.ClusterIPNone
	}
	return &service
}

func newZkPdb(z *v1beta1.ZookeeperCluster) *policyv1beta1.PodDisruptionBudget {
	pdbCount := intstr.FromInt(1)
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName(),
			Namespace: z.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(z, schema.GroupVersionKind{
					Group:   v1beta1.SchemeGroupVersion.Group,
					Version: v1beta1.SchemeGroupVersion.Version,
					Kind:    "ZookeeperCluster",
				}),
			},
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			MaxUnavailable: &pdbCount,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": z.GetName(),
				},
			},
		},
	}
}
