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
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/q8s-io/zookeeper-operator-pravega/pkg/apis/zookeeper/v1beta1"
)

const (
	externalDNSAnnotationKey = "external-dns.alpha.kubernetes.io/hostname"
	dot                      = "."
)

func headlessSvcName(z *v1beta1.ZookeeperCluster) string {
	return fmt.Sprintf("%s-headless", z.GetName())
}

var zkDataVolume = "data"

// MakeClientService returns a client service resource for the zookeeper cluster
func MakeClientService(z *v1beta1.ZookeeperCluster) *corev1.Service {
	ports := z.ZookeeperPorts()
	svcPorts := []corev1.ServicePort{
		{Name: "tcp-client", Port: ports.Client},
	}
	return makeService(z.GetClientServiceName(), svcPorts, true, z)
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

// MakeHeadlessService returns an internal headless-service for the zk
// stateful-set
func MakeHeadlessService(z *v1beta1.ZookeeperCluster) *v1.Service {
	ports := z.ZookeeperPorts()
	svcPorts := []v1.ServicePort{
		{Name: "tcp-client", Port: ports.Client},
		{Name: "tcp-quorum", Port: ports.Quorum},
		{Name: "tcp-leader-election", Port: ports.Leader},
		{Name: "tcp-metrics", Port: ports.Metrics},
	}
	return makeService(headlessSvcName(z), svcPorts, false, z)
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
