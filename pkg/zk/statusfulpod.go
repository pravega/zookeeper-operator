package zk

import (
	"reflect"
	"strings"

	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
)

// MakeStatefulPod return a zookeeper stateful set from the zk spec
func MakeStatefulPod(z *v1beta1.ZookeeperCluster) *statefulpodv1.StatefulPod {
	var extraVolumes []corev1.Volume
	persistence := z.Spec.Persistence
	var pvcs []corev1.PersistentVolumeClaim
	if strings.EqualFold(z.Spec.StorageType, "ephemeral") {
		extraVolumes = append(extraVolumes, corev1.Volume{
			Name: zkDataVolume,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &z.Spec.Ephemeral.EmptyDirVolumeSource,
			},
		})
	} else {
		pvcs = append(pvcs, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name: zkDataVolume,
				Labels: mergeLabels(
					z.Spec.Labels,
					map[string]string{"app": z.GetName()},
				),
			},
			Spec: persistence.PersistentVolumeClaimSpec,
		})
	}
	podSpec := makeZkPodSpec(z, extraVolumes)
	var port []corev1.ServicePort
	port = append(port, corev1.ServicePort{
		Port: 80,
	})
	return &statefulpodv1.StatefulPod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName(),
			Namespace: z.Namespace,
			Labels:    z.Spec.Labels,
		},
		Spec: statefulpodv1.StatefulPodSpec{
			Size: &z.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": z.GetName(),
				},
			},
			PodTemplate: podSpec,
			ServiceTemplate: &corev1.ServiceSpec{
				Ports: port,
				// Selector:  myselector,
				Selector:  map[string]string{"app": z.GetName()},
				ClusterIP: "None",
			},
		},
	}
}

func makeZkPodSpec(z *v1beta1.ZookeeperCluster, volumes []corev1.Volume) corev1.PodSpec {
	zkContainer := corev1.Container{
		Name:  "zookeeper",
		Image: z.Spec.Image.ToString(),
		Ports: z.Spec.Ports,
		Env: []corev1.EnvVar{
			{
				Name: "ENVOY_SIDECAR_STATUS",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: `metadata.annotations['sidecar.istio.io/status']`,
					},
				},
			},
		},
		ImagePullPolicy: z.Spec.Image.PullPolicy,
		// ReadinessProbe: &corev1.Probe{
		// 	InitialDelaySeconds: 10,
		// 	TimeoutSeconds:      10,
		// 	Handler: corev1.Handler{
		// 		Exec: &corev1.ExecAction{Command: []string{"/usr/local/bin/zookeeperReady.sh"}},
		// 	},
		// },
		LivenessProbe: &corev1.Probe{
			InitialDelaySeconds: 10,
			TimeoutSeconds:      10,
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{Command: []string{"/usr/local/bin/zookeeperLive.sh"}},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{Name: "conf", MountPath: "/conf"},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{"zookeeperTeardown.sh"},
				},
			},
		},
		Command: []string{"/usr/local/bin/zookeeperStart.sh"},
	}
	if z.Spec.Pod.Resources.Limits != nil || z.Spec.Pod.Resources.Requests != nil {
		zkContainer.Resources = z.Spec.Pod.Resources
	}
	volumes = append(volumes, corev1.Volume{
		Name: "conf",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: z.ConfigMapName(),
				},
			},
		},
	})

	zkContainer.Env = append(zkContainer.Env, z.Spec.Pod.Env...)
	podSpec := corev1.PodSpec{
		Containers: append(z.Spec.Containers, zkContainer),
		Affinity:   z.Spec.Pod.Affinity,
		Volumes:    append(z.Spec.Volumes, volumes...),
	}
	if reflect.DeepEqual(corev1.PodSecurityContext{}, z.Spec.Pod.SecurityContext) {
		podSpec.SecurityContext = z.Spec.Pod.SecurityContext
	}
	podSpec.NodeSelector = z.Spec.Pod.NodeSelector
	podSpec.Tolerations = z.Spec.Pod.Tolerations
	podSpec.TerminationGracePeriodSeconds = &z.Spec.Pod.TerminationGracePeriodSeconds
	podSpec.ServiceAccountName = z.Spec.Pod.ServiceAccountName

	return podSpec
}
