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
	"reflect"
	"strconv"
	"strings"

	statefulpodv1 "github.com/q8s-io/iapetos/api/v1"

	// pingcapv1 "github.com/q8s-io/statefulset-pingcap/client/apis/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
)

const (
	externalDNSAnnotationKey = "external-dns.alpha.kubernetes.io/hostname"
	dot                      = "."
)

func headlessDomain(z *v1beta1.ZookeeperCluster) string {
	// return fmt.Sprintf("%s.%s.svc.%s", headlessSvcName(z), z.GetNamespace(), z.GetKubernetesClusterDomain())
	return fmt.Sprintf("%s", z.GetKubernetesClusterDomain())
}

func headlessSvcName(z *v1beta1.ZookeeperCluster) string {
	return fmt.Sprintf("%s-headless", z.GetName())
}

var zkDataVolume = "data"

// MakeStatefulSet return a zookeeper stateful set from the zk spec
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

// MakeClientService returns a client service resource for the zookeeper cluster
func MakeClientService(z *v1beta1.ZookeeperCluster) *corev1.Service {
	ports := z.ZookeeperPorts()
	svcPorts := []corev1.ServicePort{
		{Name: "tcp-client", Port: ports.Client},
	}
	return makeService(z.GetClientServiceName(), svcPorts, true, z)
}

// MakeConfigMap returns a zookeeper config map
func MakeConfigMap(z *v1beta1.ZookeeperCluster) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.ConfigMapName(),
			Namespace: z.Namespace,
			Labels:    z.Spec.Labels,
		},
		Data: map[string]string{
			"zoo.cfg":                makeZkConfigString(z.Spec),
			"log4j.properties":       makeZkLog4JConfigString(),
			"log4j-quiet.properties": makeZkLog4JQuietConfigString(),
			"env.sh":                 makeZkEnvConfigString(z),
		},
	}
}

// MakeHeadlessService returns an internal headless-service for the zk
// stateful-set
func MakeHeadlessService(z *v1beta1.ZookeeperCluster) *corev1.Service {
	ports := z.ZookeeperPorts()
	svcPorts := []corev1.ServicePort{
		{Name: "tcp-client", Port: ports.Client},
		{Name: "tcp-quorum", Port: ports.Quorum},
		{Name: "tcp-leader-election", Port: ports.Leader},
		{Name: "tcp-metrics", Port: ports.Metrics},
	}
	return makeService(headlessSvcName(z), svcPorts, false, z)
}

func makeZkConfigString(s v1beta1.ZookeeperClusterSpec) string {
	return "4lw.commands.whitelist=cons, envi, conf, crst, srvr, stat, mntr, ruok\n" +
		"dataDir=/data\n" +
		"standaloneEnabled=false\n" +
		"reconfigEnabled=true\n" +
		"skipACL=yes\n" +
		"metricsProvider.className=org.apache.zookeeper.metrics.prometheus.PrometheusMetricsProvider\n" +
		"metricsProvider.httpPort=7000\n" +
		"metricsProvider.exportJvmInfo=true\n" +
		"initLimit=" + strconv.Itoa(s.Conf.InitLimit) + "\n" +
		"syncLimit=" + strconv.Itoa(s.Conf.SyncLimit) + "\n" +
		"tickTime=" + strconv.Itoa(s.Conf.TickTime) + "\n" +
		"globalOutstandingLimit=" + strconv.Itoa(s.Conf.GlobalOutstandingLimit) + "\n" +
		"preAllocSize=" + strconv.Itoa(s.Conf.PreAllocSize) + "\n" +
		"snapCount=" + strconv.Itoa(s.Conf.SnapCount) + "\n" +
		"commitLogCount=" + strconv.Itoa(s.Conf.CommitLogCount) + "\n" +
		"snapSizeLimitInKb=" + strconv.Itoa(s.Conf.SnapSizeLimitInKb) + "\n" +
		"maxCnxns=" + strconv.Itoa(s.Conf.MaxCnxns) + "\n" +
		"maxClientCnxns=" + strconv.Itoa(s.Conf.MaxClientCnxns) + "\n" +
		"minSessionTimeout=" + strconv.Itoa(s.Conf.MinSessionTimeout) + "\n" +
		"maxSessionTimeout=" + strconv.Itoa(s.Conf.MaxSessionTimeout) + "\n" +
		"autopurge.snapRetainCount=" + strconv.Itoa(s.Conf.AutoPurgeSnapRetainCount) + "\n" +
		"autopurge.purgeInterval=" + strconv.Itoa(s.Conf.AutoPurgePurgeInterval) + "\n" +
		"quorumListenOnAllIPs=" + strconv.FormatBool(s.Conf.QuorumListenOnAllIPs) + "\n" +
		"dynamicConfigFile=/data/zoo.cfg.dynamic\n" +
		// clientPort
		"clientPort=2181"
}

func makeZkLog4JQuietConfigString() string {
	return "log4j.rootLogger=ERROR, CONSOLE\n" +
		"log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n" +
		"log4j.appender.CONSOLE.Threshold=ERROR\n" +
		"log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n" +
		"log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n\n"
}

func makeZkLog4JConfigString() string {
	return "zookeeper.root.logger=CONSOLE\n" +
		"zookeeper.console.threshold=INFO\n" +
		"log4j.rootLogger=${zookeeper.root.logger}\n" +
		"log4j.appender.CONSOLE=org.apache.log4j.ConsoleAppender\n" +
		"log4j.appender.CONSOLE.Threshold=${zookeeper.console.threshold}\n" +
		"log4j.appender.CONSOLE.layout=org.apache.log4j.PatternLayout\n" +
		"log4j.appender.CONSOLE.layout.ConversionPattern=%d{ISO8601} [myid:%X{myid}] - %-5p [%t:%C{1}@%L] - %m%n\n"
}

func makeZkEnvConfigString(z *v1beta1.ZookeeperCluster) string {
	ports := z.ZookeeperPorts()
	return "#!/usr/bin/env bash\n\n" +
		// "DOMAIN=" + headlessDomain(z) + "\n" +
		"DOMAIN=" + z.GetName() + "-service" + "\n" +
		"QUORUM_PORT=" + strconv.Itoa(int(ports.Quorum)) + "\n" +
		"LEADER_PORT=" + strconv.Itoa(int(ports.Leader)) + "\n" +
		"CLIENT_HOST=" + z.GetClientServiceName() + "\n" +
		"CLIENT_PORT=" + strconv.Itoa(int(ports.Client)) + "\n" +
		"CLUSTER_NAME=" + z.GetName() + "\n" +
		"CLUSTER_SIZE=" + fmt.Sprint(z.Spec.Replicas) + "\n"
}

func makeService(name string, ports []corev1.ServicePort, clusterIP bool, z *v1beta1.ZookeeperCluster) *corev1.Service {
	var dnsName string
	var annotationMap map[string]string
	if !clusterIP && z.Spec.DomainName != "" {
		domainName := strings.TrimSpace(z.Spec.DomainName)
		if strings.HasSuffix(domainName, dot) {
			dnsName = name + dot + domainName
		} else {
			dnsName = name + dot + domainName + dot
		}
		annotationMap = map[string]string{externalDNSAnnotationKey: dnsName}
	} else {
		annotationMap = map[string]string{}
	}
	service := corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: z.Namespace,
			Labels: mergeLabels(
				z.Spec.Labels,
				map[string]string{"app": z.GetName(), "headless": strconv.FormatBool(!clusterIP)},
			),
			Annotations: annotationMap,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: map[string]string{"app": z.GetName()},
		},
	}
	if !clusterIP {
		service.Spec.ClusterIP = corev1.ClusterIPNone
	}
	return &service
}

// MakePodDisruptionBudget returns a pdb for the zookeeper cluster
func MakePodDisruptionBudget(z *v1beta1.ZookeeperCluster) *policyv1beta1.PodDisruptionBudget {
	pdbCount := intstr.FromInt(1)
	return &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PodDisruptionBudget",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.GetName(),
			Namespace: z.Namespace,
			Labels:    z.Spec.Labels,
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

//MakeServiceAccount returns the service account for zookeeper Cluster
func MakeServiceAccount(z *v1beta1.ZookeeperCluster) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      z.Spec.Pod.ServiceAccountName,
			Namespace: z.Namespace,
		},
	}
}

// MergeLabels merges label maps
func mergeLabels(l ...map[string]string) map[string]string {
	res := make(map[string]string)

	for _, v := range l {
		for lKey, lValue := range v {
			res[lKey] = lValue
		}
	}
	return res
}

func GetPVCName() string {
	return zkDataVolume
}
