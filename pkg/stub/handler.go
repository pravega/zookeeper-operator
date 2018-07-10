package stub

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
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
}

type zkPorts struct {
	Client int32
	Quorum int32
	Leader int32
}

func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1beta1.ZookeeperCluster:
		o.WithDefaults()
		var ports zkPorts
		for _, p := range o.Spec.Ports {
			if p.Name == "client" {
				ports.Client = p.ContainerPort
			} else if p.Name == "quorum" {
				ports.Quorum = p.ContainerPort
			} else if p.Name == "leader-election" {
				ports.Leader = p.ContainerPort
			}
		}
		configMapName := fmt.Sprintf("%s-configmap", o.GetName())
		err := sdk.Create(newZkConfigMap(configMapName, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper configmap : %v", err)
			return err
		}
		err = sdk.Create(newZkSts(configMapName, ports, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper statefulset : %v", err)
			return err
		}
		err = sdk.Create(newZkPdb(o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper pod-disruption-budget : %v", err)
			return err
		}
		err = sdk.Create(newZkClientSvc(ports, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper client service : %v", err)
			return err
		}
		err = sdk.Create(newZkHeadlessSvc(ports, o))
		if err != nil && !k8serrors.IsAlreadyExists(err) {
			logrus.Errorf("Failed to create zookeeper headless service : %v", err)
			return err
		}
	}
	return nil
}

// newZkSts creates a new Zookeeper StatefulSet
func newZkSts(configMapName string, ports zkPorts, z *v1beta1.ZookeeperCluster) *appsv1.StatefulSet {
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
			ServiceName: fmt.Sprintf("%s-headless", z.GetName()),
			Replicas:    &z.Spec.Size,
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
						"app":  z.GetName(),
						"kind": "ZookeeperMember",
					},
				},
				Spec: newZkPodSpec(configMapName, ports, z),
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

func newZkPodSpec(configMapName string, ports zkPorts, z *v1beta1.ZookeeperCluster) v1.PodSpec {
	probe := v1.Probe{
		InitialDelaySeconds: 10,
		TimeoutSeconds:      10,
		Handler: v1.Handler{
			Exec: &v1.ExecAction{
				Command: []string{"zookeeperReady.sh", strconv.Itoa(int(ports.Client))},
			},
		},
	}
	zkContainer := v1.Container{
		Name:            "zookeeper",
		Image:           z.Spec.Image.ToString(),
		Ports:           z.Spec.Ports,
		ImagePullPolicy: z.Spec.Image.PullPolicy,
		ReadinessProbe:  &probe,
		LivenessProbe:   &probe,
		VolumeMounts: []v1.VolumeMount{
			{Name: "data", MountPath: "/data"},
			{Name: "conf", MountPath: "/conf"},
		},
	}
	if z.Spec.Pod.Resources.Limits != nil || z.Spec.Pod.Resources.Requests != nil {
		zkContainer.Resources = z.Spec.Pod.Resources
	}
	zkContainer.Env = z.Spec.Pod.Env
	podSpec := v1.PodSpec{
		InitContainers: []v1.Container{
			{
				Name:  "zookeeper-init",
				Image: "spiegela/zookeeper-init:latest",
				VolumeMounts: []v1.VolumeMount{
					{Name: "data", MountPath: "/data"},
					{Name: "conf", MountPath: "/conf"},
				},
				Args: []string{
					fmt.Sprintf("%s-headless.%s.svc.cluster.local", z.GetName(), z.GetNamespace()),
					strconv.Itoa(int(ports.Client)),
					strconv.Itoa(int(ports.Quorum)),
					strconv.Itoa(int(ports.Leader)),
				},
			},
		},
		Containers: []v1.Container{zkContainer},
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
func newZkClientSvc(ports zkPorts, z *v1beta1.ZookeeperCluster) *v1.Service {
	name := fmt.Sprintf("%s-client", z.GetName())
	svcPorts := []v1.ServicePort{
		{Name: "client", Port: ports.Client},
	}
	return newSvc(name, svcPorts, true, z)
}

// newZkInternalSvc creates a new Zookeeper Service
func newZkHeadlessSvc(ports zkPorts, z *v1beta1.ZookeeperCluster) *v1.Service {
	name := fmt.Sprintf("%s-headless", z.GetName())
	svcPorts := []v1.ServicePort{
		{Name: "quorum", Port: ports.Quorum},
		{Name: "leader-election", Port: ports.Leader},
	}
	return newSvc(name, svcPorts, false, z)
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
			"initLimit=" + strconv.Itoa(z.Spec.Conf.InitLimit) + "\n" +
			"syncLimit=" + strconv.Itoa(z.Spec.Conf.SyncLimit) + "\n" +
			"tickTime=" + strconv.Itoa(z.Spec.Conf.TickTime) + "\n" +
			"dynamicConfigFile=/data/zoo.cfg.dynamic\n",
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
			Labels: map[string]string{"app": z.GetName()},
		},
		Spec: v1.ServiceSpec{
			Ports:    ports,
			Selector: map[string]string{"app": z.GetName()},
		},
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
